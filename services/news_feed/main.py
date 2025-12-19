import os
from crewai import Agent, Task, Crew, Process, LLM
import time

import firestore_pb2
import firestore_pb2_grpc

from tools import DuckDuckGoTool



# tools and llm setup
ddgs = DuckDuckGoTool()

llm = LLM(model="ollama/mistral", api_base="http://localhost:11434") # note: llm may refuse if consider harmful



# Creating Specialized Agents
PatientResearcher = Agent(
    role="Senior Research Analyst",
    goal="Compile an report of multiple news articles and brief descriptions of dementia-related news, focusing on articles that patients suffering from dementia will find useful, informative, or uplifting. minimizing technical words and providing simple definitions to any technical words used.",
    verbose=True,
    memory=True,
	backstory="""
	You are a distinguished research analyst with a focus on dementia research and science communication.
	You are a passionate journalist known for making complex medical subjects accessible to patients and general audiences.
	With a background in science communication and creative writing, you provide informative and easy to understand news relevant to patients with dementia.
	You strive to provide the most up to date and relevant news for patients suffering with dementia, you excel at explaining complex and confusing concepts to patients who have little medical knowledge.
	""",
    tools=[ddgs],
    llm=llm,
    allow_delegation=True
)


# Defining Tasks
PatientTask = Task(
	description="""
	Compile an report of multiple news articles.

	Research multiple news articles that a patient suffering from dementia might find useful, informative, funny, or uplifting.
	Including:
	1) a quick descriptions of the news
	2) then a detailed descriptions of the news
	3) research any terminology a dementia patient would not know and provide a simple and concise explanation, do not explain dementia.
	4) Explains how may affect a person living with dementia

	if an article is not in english, ignore it and find a diffrent article.
	""",
    expected_output="A detailed news report containing a paragraph for each previously mentioned news article, in plaintext for patients suffering from dementia. without links.",
    tools=[ddgs],
    output_file='output-patient.md',
    agent=PatientResearcher,
    llm=llm
)
# tasks) research, compile using research (what second for?)

DoctorResearcher = Agent(
    role="Senior Research Analyst",
    goal="Compile an report of multiple news articles and brief descriptions of dementia-related news in medicine as well as advancements in medical research, focusing on articles that doctors researching dementia will find useful and informative.",
    verbose=True,
    memory=True,
	backstory="""
	You are a distinguished research analyst with a focus on dementia research and scientific research.
	You are an inteligent researcher that has written and reviewed many medical papers.
	You strive to provide the most up to date and relevant news for doctors studying dementia, you provide detailed explanations of any informatio provided.
	""",
    tools=[ddgs],
    llm=llm,
    allow_delegation=True
)

DoctorTask = Task(
	description="""
	Compile an report of multiple news articles.

	Research multiple news articles that a doctor researching dementia might find useful and informative.
	Including:
	1) a quick descriptions of the news
	2) then a detailed descriptions of the news
	3) Explains how may affect a person living with dementia
	4) Explain applications to medical practice.

	if an article is not in english, ignore it and find a diffrent article.
	""",
    expected_output="A detailed news report containing a paragraph for each previously mentioned news article, in plaintext for patients suffering from dementia. without links.",
    tools=[ddgs],
    output_file='output-doctor.md',
    agent=PatientResearcher,
    llm=llm
)







def main():
	sleep(60*60*24)
	# Initialize the crew
	result = Crew(
		agents=[PatientResearcher],
		tasks=[PatientTask],
		process=Process.sequential,
		).kickoff()
	# with open("output.md","w") as f:
		# f.write(str(result))
	result = Crew(
		agents=[DoctorResearcher],
		tasks=[DoctorTask],
		process=Process.sequential,
		).kickoff()
	with open("output.md","w") as f:
		f.write(str(result))
	port = "50052"
	channel = grpc.insecure_channel('localhost:'+port)
	stub = firestore_pb2_grpc.firestore(channel)
	req = firestore_pb2.NewsSet(Type="Patient",Content=str(result))
	resp = stub.SetNews(req)

main()



# os.system("notify-send researchDone")
