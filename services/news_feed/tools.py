from crewai.tools import BaseTool
from langchain_community.tools import DuckDuckGoSearchRun


class DuckDuckGoTool(BaseTool):
	name: str = "DuckDuckGo Search Tool"
	description: str = "Search the web for a given query. use the format '{ query: QUERY_STRING }'.the query should be a string, do not include any aditional information or dictionarys"

	def _run(self, query: str) -> str:
		# Ensure the DuckDuckGoSearchRun is invoked properly.
		duckduckgo_tool = DuckDuckGoSearchRun()
		response = duckduckgo_tool.invoke(query)
		with open("tool_log.txt","a") as f:
			f.write(str(query)+"\n")
			f.write(str(response)+"\n---\n\n")
		return response

	def _get_tool(self):
		# Create an instance of the tool when needed
		return DuckDuckGoTool()
