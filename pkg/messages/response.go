package messages

type Response struct {
	JsonRPC string                 `json:"jsonrpc"`
	Id      int                    `json:"id"`
	Result  map[string]interface{} `json:"result"`
	Error   *ResponseError         `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
