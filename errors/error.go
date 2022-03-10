package errors

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrResponse struct {
	RequestID      string `json:"requestId"`
	HttpStatusCode int    `json:"code"`
	StatusText     string `json:"status"`
}

func (e *ErrResponse) Render(rw http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HttpStatusCode)
	return nil
}

func NoErr(tId string) render.Renderer {
	return &ErrResponse{
		RequestID:      tId,
		HttpStatusCode: 200,
		StatusText:     "complete",
	}
}

func GenericErr(tId string, code int, msg string) render.Renderer {
	return &ErrResponse{
		RequestID:      tId,
		HttpStatusCode: code,
		StatusText:     msg,
	}
}

func ErrConflict(tId string) render.Renderer {
	return &ErrResponse{
		RequestID:      tId,
		HttpStatusCode: 409,
		StatusText:     "conflict",
	}
}

func ErrFobiddenRequest(tId string) render.Renderer {
	return &ErrResponse{
		RequestID:      tId,
		HttpStatusCode: 401,
		StatusText:     "forbidden",
	}
}

func ErrInvalidRequest(tId string) render.Renderer {
	return &ErrResponse{
		RequestID:      tId,
		HttpStatusCode: 400,
		StatusText:     "invalid request",
	}
}

func ErrInternal(tId string) render.Renderer {
	return &ErrResponse{
		RequestID:      tId,
		HttpStatusCode: 500,
		StatusText:     "internal error",
	}
}

func ErrRender(tId string) render.Renderer {
	return &ErrResponse{
		RequestID:      tId,
		HttpStatusCode: 422,
		StatusText:     "error rendering response",
	}
}
