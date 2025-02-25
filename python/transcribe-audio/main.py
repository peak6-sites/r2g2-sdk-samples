import base64
import os
import sys

import requests

def transcribe_audio(inline_data=None, file_data_uri=None, mime_type="audio/wav"):
    url = "https://api-proxy-prod.prod.gcp.minisme.ai/ai.audio.v0.Transcriber/Transcribe"

    headers = {
        "Authorization": f"Bearer {os.getenv('R2G2_TOKEN')}",
        "Content-Type": "application/json"
    }

    # Construct request payload
    if inline_data:
        data = {
            "inline_data": {
                "mime_type": mime_type,
                "data": base64.b64encode(inline_data).decode("ascii"),
            }
        }
    elif file_data_uri:
        data = {
            "file_data": {
                "mime_type": mime_type,
                "uri": file_data_uri
            }
        }
    else:
        raise ValueError("Either inline_data or file_data_uri must be provided.")

    response = requests.post(url, headers=headers, json=data)
    response.raise_for_status()

    transcription = response.json()
    return transcription["text"]

if __name__ == "__main__":
    with open(sys.argv[1], "rb") as f:
        data = f.read()
    print(transcribe_audio(inline_data=data, mime_type="audio/mpeg"))