from langchain.agents import load_tools, initialize_agent, AgentType
from langchain.chains import LLMChain
from langchain.llms import OpenAI
from langchain.prompts import PromptTemplate
from langchain import ConversationChain

# llm = OpenAI(temperature=0.9)
# text = "What would be a good company name for a company that makes colorful socks?"
# print(llm(text))

# prompt = PromptTemplate(
#     input_variables=["product"],
#     template="What is a good name for a company that makes {product}?",
# )

# print(prompt.format(product="colorful socks"))
#
# chain = LLMChain(llm=llm, prompt=prompt)
# print(chain.run("colorful socks"))

# llm = OpenAI(temperature=0.9)
# tools = load_tools(["serpapi", "llm-math"], llm=llm)
#
# agent = initialize_agent(tools, llm, agent=AgentType.ZERO_SHOT_REACT_DESCRIPTION, verbose=True)
#
# print(agent.run("What was the high temperature in SF yesterday in Fahrenheit? What is that number raised to the .023 power?"))

# llm = OpenAI(temperature=0)
# conversation = ConversationChain(llm=llm, verbose=True)
# output = conversation.predict(input="Hi there!")
# print(output)

from langchain.document_loaders import BigQueryLoader
from langchain.indexes import VectorstoreIndexCreator
from langchain.embeddings import HuggingFaceEmbeddings
from langchain.vectorstores import Chroma

QUERY = """
SELECT
    CONCAT(questions.id, "/", answers.id) AS id,
    CONCAT(questions.body, "\\n", answers.body) AS body
FROM `bigquery-public-data.stackoverflow.posts_questions` AS questions
INNER JOIN `bigquery-public-data.stackoverflow.posts_answers` AS answers
    ON questions.id = answers.parent_id
WHERE questions.creation_date >= "2023-01-01"
"""

llm = OpenAI(temperature=0.9)
loader = BigQueryLoader(QUERY, page_content_columns=["body"], metadata_columns=["id"])
embedding = HuggingFaceEmbeddings(model_name="sentence-transformers/all-mpnet-base-v2",
        model_kwargs={"device": "cpu"})
# vectordb = Chroma(embedding_function=embedding).from_loaders([loader])
index = VectorstoreIndexCreator(embedding=embedding).from_loaders([loader])
result = index.query_with_sources("How can I query Spanner in Go", llm=llm)
print(result)
