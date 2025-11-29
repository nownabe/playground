import logging
from typing import AsyncGenerator
from typing_extensions import override

from google.adk.agents import LlmAgent, BaseAgent, LoopAgent, SequentialAgent
from google.adk.agents.invocation_context import InvocationContext
from google.adk.events import Event
from google.genai import types


GEMINI_2_5_FLASH = "gemini-2.5-flash"


logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class StoryFlowAgent(BaseAgent):
    """
    ストーリー生成と洗練のためのカスタムエージェント。

    このエージェントはLLMエージェントのシーケンスを調整してストーリーを生成し、
    批評し、修正し、文法とトーンをチェックし、もしトーンがネガティブであれば
    ストーリーを再生成する可能性があります。
    """

    story_generator: LlmAgent
    critic: LlmAgent
    reviser: LlmAgent
    grammar_check: LlmAgent
    tone_check: LlmAgent

    loop_agent: LoopAgent
    sequential_agent: SequentialAgent

    model_config = {"arbitrary_types_allowed": True}

    def __init__(
        self,
        name: str,
        story_generator: LlmAgent,
        critic: LlmAgent,
        reviser: LlmAgent,
        grammar_check: LlmAgent,
        tone_check: LlmAgent,
    ):
        """
        StoryFlowAgentを初期化します。

        Args:
            name: エージェントの名前
            story_generator: 初期ストーリーを生成するLlmAgent
            critic: ストーリーを批評するLlmAgent
            reviser: 批評に基づいてストーリーを修正するLlmAgent
            grammar_check: 文法をチェックするLlmAgent
            tone_check: トーンを分析するLlmAgent
        """

        loop_agent = LoopAgent(
            name="CriticReviserLoop",
            sub_agents=[critic, reviser],
            max_iterations=2,
        )
        sequential_agent = SequentialAgent(
            name="PostProcessing",
            sub_agents=[grammar_check, tone_check],
        )
        sub_agents = [story_generator, loop_agent, sequential_agent]

        super().__init__(
            name=name,
            story_generator=story_generator,
            critic=critic,
            reviser=reviser,
            grammar_check=grammar_check,
            tone_check=tone_check,
            loop_agent=loop_agent,
            sequential_agent=sequential_agent,
            sub_agents=sub_agents,
        )

    @override
    async def _run_async_impl(
        self, ctx: InvocationContext
    ) -> AsyncGenerator[Event, None]:
        """
        ストーリーワークフローのカスタムオーケストレーションロジックを実装します。
        Pydanticによって割り当てられたインスタンス属性(例: self.story_generator)を使用します。
        """

        logger.info(f"[{self.name}] ストーリー生成ワークフローを開始します")

        logger.info(f"[{self.name}] StoryGenerator を実行中...")
        async for event in self.story_generator.run_async(ctx):
            logger.info(
                f"[{self.name}] StoryGenerator からのイベント: {event.model_dump_json(indent=2, exclude_none=True)}"
            )
            yield event

        if (
            "current_story" not in ctx.session.state
            or not ctx.session.state["current_story"]
        ):
            logger.error(
                f"[{self.name}] 初期ストーリーの生成に失敗しました。ワークフローを中断します"
            )
            return

        logger.info(
            f"[{self.name}] ジェネレーター後のストーリーの状態: {ctx.session.state.get('current_story')}"
        )

        # 2. 批評家-修正者ループ
        logger.info(f"[{self.name}] CriticReviserLoop を実行中...")
        # 初期化時に割り当てられた loop_agent インスタンス属性を使用
        async for event in self.loop_agent.run_async(ctx):
            logger.info(
                f"[{self.name}] CriticReviserLoop からのイベント: {event.model_dump_json(indent=2, exclude_none=True)}"
            )
            yield event

        logger.info(
            f"[{self.name}] ループ後のストーリーの状態: {ctx.session.state.get('current_story')}"
        )

        # 3. 逐次後処理（文法とトーンのチェック）
        logger.info(f"[{self.name}] PostProcessing を実行中...")
        # 初期化時に割り当てられた sequential_agent インスタンス属性を使用
        async for event in self.sequential_agent.run_async(ctx):
            logger.info(
                f"[{self.name}] PostProcessing からのイベント: {event.model_dump_json(indent=2, exclude_none=True)}"
            )
            yield event

        # 4. トーンに基づく条件ロジック
        tone_check_result = ctx.session.state.get("tone_check_result")
        logger.info(f"[{self.name}] トーンチェック結果: {tone_check_result}")

        if tone_check_result == "negative":
            logger.info(
                f"[{self.name}] トーンがネガティブです。ストーリーを再生成します..."
            )
            async for event in self.story_generator.run_async(ctx):
                logger.info(
                    f"[{self.name}] StoryGenerator からのイベント (再生成): {event.model_dump_json(indent=2, exclude_none=True)}"
                )
                yield event
        else:
            logger.info(
                f"[{self.name}] トーンはネガティブではありません。現在のストーリーを維持します。"
            )
            pass

        generated_story = ctx.session.state.get("current_story")
        yield Event(
            invocation_id=ctx.invocation_id,
            content=types.Content(
                role="model", parts=[types.Part.from_text(text=generated_story)]
            ),
            author=ctx.agent.name,
        )

        logger.info(f"[{self.name}] ワークフローが完了しました。")


# --- 個々のLLMエージェントを定義 ---
story_generator = LlmAgent(
    name="StoryGenerator",
    model=GEMINI_2_5_FLASH,
    instruction="""あなたは物語作家です。ユーザによって提供されたトピックに基づいて、猫についての短い物語（約200語）を書いてください。""",
    input_schema=None,
    output_key="current_story",  # Key for storing output in session state
)

critic = LlmAgent(
    name="Critic",
    model=GEMINI_2_5_FLASH,
    instruction="""あなたは物語の批評家です。Session Stateの 'current_story' キーで提供された物語をレビューしてください。
物語を改善する方法について、1〜2文の建設的な批判を提供してください。プロットまたはキャラクターに焦点を当ててください。""",
    input_schema=None,
    output_key="criticism",  # Key for storing criticism in session state
)

reviser = LlmAgent(
    name="Reviser",
    model=GEMINI_2_5_FLASH,
    instruction="""あなたは物語の修正者です。Session Stateの 'current_story' キーで提供された物語を、
セッション状態の 'criticism' キーにある批判に基づいて修正してください。修正された物語のみを出力してください。""",
    input_schema=None,
    output_key="current_story",  # Overwrites the original story
)

grammar_check = LlmAgent(
    name="GrammarCheck",
    model=GEMINI_2_5_FLASH,
    instruction="""あなたは文法チェッカーです。Session Stateの 'current_story' キーで提供された物語の文法をチェックしてください。
提案された修正点をリストとしてのみ出力するか、エラーがない場合は「文法は良好です！」と出力してください。""",
    input_schema=None,
    output_key="grammar_suggestions",
)

tone_check = LlmAgent(
    name="ToneCheck",
    model=GEMINI_2_5_FLASH,
    instruction="""あなたはトーンアナライザーです。Session Stateの 'current_story' キーで提供された物語のトーンを分析してください。
トーンが一般的にポジティブな場合は「positive」、一般的にネガティブな場合は「negative」、
それ以外の場合は「neutral」という単語のみを出力してください。""",
    input_schema=None,
    output_key="tone_check_result",  # This agent's output determines the conditional flow
)

# --- カスタムエージェントインスタンスを作成 ---
root_agent = StoryFlowAgent(
    name="StoryFlowAgent",
    story_generator=story_generator,
    critic=critic,
    reviser=reviser,
    grammar_check=grammar_check,
    tone_check=tone_check,
)
