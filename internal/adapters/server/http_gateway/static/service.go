package static

import (
	"net/http"

	httpgw "github.com/sinhnguyen1411/stock-trading-be/internal/adapters/server/http_gateway"
)

type Service struct {
	dir string
}

func New(dir string) httpgw.HTTPService {
	return &Service{dir: dir}
}

func (s *Service) HTTPRegister(mux *http.ServeMux) error {
	mux.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir(s.dir))))
	return nil
}
