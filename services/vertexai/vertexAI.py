from google import genai
from google.genai import types
import base64
import os

os.environ["GOOGLE_APPLICATION_CREDENTIALS"]="./copper-actor-475117-i7-92a1502a7bf4.json"

def FTproompt(proompt):
  client = genai.Client(
      vertexai=True,
      api_key=os.environ.get("GOOGLE_CLOUD_API_KEY"),
	  project='copper-actor-475117-i7', location='us-central1'
  )


  model = "projects/585981786057/locations/us-central1/endpoints/3968791674062110720"
  contents = [
    types.Content(
      role="user",
      parts=[ types.Part.from_text(text=proompt) ]
    )
  ]

  generate_content_config = types.GenerateContentConfig(
    temperature = 1,
    top_p = 0.95,
    max_output_tokens = 65535,
    safety_settings = [types.SafetySetting(
      category="HARM_CATEGORY_HATE_SPEECH",
      threshold="OFF"
    ),types.SafetySetting(
      category="HARM_CATEGORY_DANGEROUS_CONTENT",
      threshold="OFF"
    ),types.SafetySetting(
      category="HARM_CATEGORY_SEXUALLY_EXPLICIT",
      threshold="OFF"
    ),types.SafetySetting(
      category="HARM_CATEGORY_HARASSMENT",
      threshold="OFF"
    )],
    thinking_config=types.ThinkingConfig( thinking_budget=0, ),
  )

  out = ""
  for chunk in client.models.generate_content_stream(
      model = model,
      contents = contents,
      config = generate_content_config,
    ):
    txt = chunk.text
    out += txt
  return out


