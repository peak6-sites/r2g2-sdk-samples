"""
This example demonstrates how to authenticate using a Service Account and retrieve an access token that can be used
to call the API.
In this example, the service account credentials are stored in a JSON file that was previously retrieved by calling
the `r2g2 iam service-account create` command. That JSON file is provided to the application using the `R2G2_CREDENTIALS`
environment variable.
"""
import json
import os

import requests

def main():
    creds = load_credentials()
    token = get_token(creds)
    list_assistants(token)


def load_credentials():
    """
    Loads credentials from the R2G2_CREDENTIALS environment variable.

    :return: credentials
    """
    with open(os.getenv("R2G2_CREDENTIALS")) as f:
        return json.load(f)


def get_token(creds):
    """
    Retrieves a token using the provided client credentials.

    This example only retrieves the access token once and does not handle token expiration. Use a library like requests-oauthlib
    to automatically handle token expiration.

    :param creds: client credentials
    :return: access token
    """
    headers = {
        "Content-Type": "application/json"
    }
    data = {
        "client_id": creds["client_id"],
        "client_secret": creds["client_secret"],
        "audience": creds["audience"],
        "grant_type": "client_credentials",
    }
    response = requests.post(creds["token_uri"], headers=headers, json=data)
    response.raise_for_status()
    return response.json()["access_token"]


def list_assistants(token):
    """
    Retrieves the first page of Assistants from the API and prints their IDs.

    :param token: access token
    :return: None
    """
    headers = {
        "Authorization": f"Bearer {token}"
    }
    response = requests.post("https://api-proxy-prod.prod.gcp.minisme.ai/ai.assistants.v0.Assistants/ListAssistants",
                            headers=headers, json={})
    response.raise_for_status()

    response_body = response.json()
    for assistant in response_body["assistants"]:
        print(f"{assistant["displayName"]} {assistant["id"]}")


if __name__ == "__main__":
    main()
