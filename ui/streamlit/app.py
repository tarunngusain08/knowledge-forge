import os
from typing import Any

import requests
import streamlit as st


API_BASE_URL = os.getenv("API_BASE_URL", "http://localhost:8080")


def api(method: str, path: str, token: str | None = None, **kwargs: Any) -> requests.Response:
    headers = kwargs.pop("headers", {})
    if token:
        headers["Authorization"] = f"Bearer {token}"
    return requests.request(method, f"{API_BASE_URL}{path}", headers=headers, timeout=60, **kwargs)


def login() -> None:
    st.subheader("Login")
    email = st.text_input("Email", value="admin@example.com")
    password = st.text_input("Password", type="password")
    if st.button("Login", type="primary"):
        resp = api("POST", "/auth/login", json={"email": email, "password": password})
        if resp.ok:
            payload = resp.json()
            st.session_state["token"] = payload["access_token"]
            st.session_state["user"] = payload["user"]
            st.rerun()
        else:
            st.error(resp.json().get("error", "Login failed"))


def upload_tab(token: str) -> None:
    st.subheader("Upload Documents")
    file = st.file_uploader("PDF, Markdown, or TXT", type=["pdf", "md", "txt"])
    if file and st.button("Upload", type="primary"):
        resp = api("POST", "/documents", token, files={"file": (file.name, file.getvalue())})
        if resp.ok:
            st.success("Upload accepted and indexing job created.")
            st.json(resp.json())
        else:
            st.error(resp.json().get("error", "Upload failed"))


def documents_tab(token: str) -> None:
    st.subheader("Documents")
    resp = api("GET", "/documents", token)
    if not resp.ok:
        st.error(resp.text)
        return
    docs = resp.json()
    st.dataframe(docs, use_container_width=True)
    doc_id = st.text_input("Document ID to delete")
    if doc_id and st.button("Delete document"):
        delete_resp = api("DELETE", f"/documents/{doc_id}", token)
        if delete_resp.ok:
            st.success("Deleted")
        else:
            st.error(delete_resp.text)


def chat_tab(token: str) -> None:
    st.subheader("Chat")
    if "session_id" not in st.session_state:
        if st.button("Create chat session"):
            resp = api("POST", "/chat/sessions", token, json={"title": "Demo chat"})
            if resp.ok:
                st.session_state["session_id"] = resp.json()["id"]
                st.rerun()
            else:
                st.error(resp.text)
        return
    st.caption(f"Session: {st.session_state['session_id']}")
    question = st.text_area("Question")
    top_k = st.selectbox("Top K", [5, 8], index=0)
    reranker = st.toggle("Reranker", value=True)
    if st.button("Ask", type="primary") and question:
        resp = api(
            "POST",
            f"/chat/sessions/{st.session_state['session_id']}/messages",
            token,
            json={"question": question, "top_k": top_k, "reranker_enabled": reranker},
        )
        if resp.ok:
            payload = resp.json()
            st.markdown(payload["answer"])
            with st.expander("Citations"):
                st.json(payload.get("citations", []))
            with st.expander("Retrieval"):
                st.json(payload.get("retrieval", {}))
        else:
            st.error(resp.text)


def debug_tab(token: str) -> None:
    st.subheader("Retrieval Debug")
    question = st.text_input("Debug question")
    top_k = st.selectbox("Debug Top K", [5, 8], index=0)
    reranker = st.toggle("Debug reranker", value=True)
    if st.button("Run debug") and question:
        resp = api("GET", "/debug/retrieval", token, params={"question": question, "top_k": top_k, "reranker": str(reranker).lower()})
        if resp.ok:
            st.json(resp.json())
        else:
            st.error(resp.text)


def eval_tab(token: str) -> None:
    st.subheader("Evaluation")
    sample = [
        {"question": "What is the PTO policy?", "expected_vector_ids": []},
        {"question": "What does the handbook say about remote work?", "expected_vector_ids": []},
    ]
    dataset = st.text_area("Questions JSON", value=str(sample).replace("'", '"'), height=160)
    top_k = st.selectbox("Eval Top K", [5, 8], index=0)
    reranker = st.toggle("Eval reranker", value=True)
    if st.button("Run evaluation"):
        import json

        questions = json.loads(dataset)
        resp = api(
            "POST",
            "/eval/runs",
            token,
            json={"name": "streamlit-eval", "top_k": top_k, "reranker_enabled": reranker, "questions": questions},
        )
        if resp.ok:
            st.json(resp.json())
        else:
            st.error(resp.text)


def main() -> None:
    st.set_page_config(page_title="RAG-bot", layout="wide")
    st.title("RAG-bot")
    token = st.session_state.get("token")
    if not token:
        login()
        return
    st.caption(f"API: {API_BASE_URL}")
    tabs = st.tabs(["Upload", "Documents", "Chat", "Retrieval Debug", "Evaluation"])
    with tabs[0]:
        upload_tab(token)
    with tabs[1]:
        documents_tab(token)
    with tabs[2]:
        chat_tab(token)
    with tabs[3]:
        debug_tab(token)
    with tabs[4]:
        eval_tab(token)


if __name__ == "__main__":
    main()

