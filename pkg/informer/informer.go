package informer

type InformerClient struct {
	info string
}

func NewInformerClient() *InformerClient {
	return &InformerClient{}
}

func (o *InformerClient) AppendInfo(s string) {
	if o.info != "" {
		o.info += "\n"
	}
	o.info += s
}

func (o *InformerClient) GetInfo() string {
	return o.info
}
