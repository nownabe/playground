import os

from google.adk.agents import Agent
from google.adk.agents.llm_agent import Agent
from google.adk.tools.retrieval import VertexAiRagRetrieval
from vertexai import rag

RAG_CORPUS_ID = "projects/tmp-adk-rag/locations/us-west1/ragCorpora/4611686018427387904"

syllabus_retrieval_tool = VertexAiRagRetrieval(
    name="SyllabusRetrievalTool",
    description="Use this tool to retrieve syllabus information for the question from the RAG corpus.",
    rag_resources=[rag.RagResource(rag_corpus=RAG_CORPUS_ID)],
)

root_agent = Agent(
    model="gemini-2.5-flash",
    name="SyllabusAgent",
    description="大学のシラバスに関する質問に答えるエージェント",
    instruction="""
        あなたは大学のシラバスに関する質問に答えるAIアシスタントです。
        ユーザーからの質問に対して、提供されたシラバス情報に基づいて正確に回答してください。

        [タスク]
        - シラバスに関する質問には、SyllabusRetrievalTool を使用して回答を生成してください。
        - 質問がシラバスに関連しない場合は、その旨を伝えてください。

        [制約]
        - シラバス情報にない内容については、推測で回答せず「シラバスにはその情報がありません。」と答えてください。
        - 常に丁寧な言葉遣いを心がけてください。
    """,
    tools=[syllabus_retrieval_tool],
)
