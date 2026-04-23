def fetch_user(user_id):
    return {"id": user_id, "name": "Ada"}


def normalize_user(user):
    return {"id": user["id"], "name": user["name"].strip()}
