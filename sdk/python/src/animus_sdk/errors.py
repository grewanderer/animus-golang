class AnimusAPIError(RuntimeError):
    def __init__(self, status: int, code: str, request_id: str | None = None, body: object | None = None) -> None:
        self.status = status
        self.code = code
        self.request_id = request_id
        self.body = body
        super().__init__(f"{code} (status={status}, request_id={request_id})")

