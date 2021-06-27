package state

type ClientStateHandler struct {
}

func (h *ClientStateHandler) Connect(url string) error {
	return nil
}

func (h *ClientStateHandler) Disconnect() {

}

func (h *ClientStateHandler) SendUpdate() {

}
