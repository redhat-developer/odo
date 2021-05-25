package v1alpha1

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type conditionsBuilder struct {
	cndType string
	status  v1.ConditionStatus
	reason  string
	message string
}

func Conditions() *conditionsBuilder {
	return &conditionsBuilder{}
}

func (builder *conditionsBuilder) Build() *v1.Condition {
	return &v1.Condition{
		Type:    builder.cndType,
		Status:  builder.status,
		Reason:  builder.reason,
		Message: builder.message,
	}
}

func (builder *conditionsBuilder) NotCollectionReady() *conditionsBuilder {
	builder.status = v1.ConditionFalse
	builder.cndType = CollectionReady
	return builder
}

func (builder *conditionsBuilder) CollectionReady() *conditionsBuilder {
	builder.status = v1.ConditionTrue
	builder.cndType = CollectionReady
	return builder
}

func (builder *conditionsBuilder) NotInjectionReady() *conditionsBuilder {
	builder.status = v1.ConditionFalse
	builder.cndType = InjectionReady
	return builder
}

func (builder *conditionsBuilder) InjectionReady() *conditionsBuilder {
	builder.status = v1.ConditionTrue
	builder.cndType = InjectionReady
	return builder
}

func (builder *conditionsBuilder) NotBindingReady() *conditionsBuilder {
	builder.status = v1.ConditionFalse
	builder.cndType = BindingReady
	return builder
}

func (builder *conditionsBuilder) BindingReady() *conditionsBuilder {
	builder.status = v1.ConditionTrue
	builder.cndType = BindingReady
	return builder
}

func (builder *conditionsBuilder) Reason(r string) *conditionsBuilder {
	builder.reason = r
	return builder
}

func (builder *conditionsBuilder) BindingInjected() *conditionsBuilder {
	builder.reason = BindingInjectedReason
	return builder
}

func (builder *conditionsBuilder) DataCollected() *conditionsBuilder {
	builder.reason = DataCollectedReason
	return builder
}

func (builder *conditionsBuilder) ServiceNotFound() *conditionsBuilder {
	builder.reason = ServiceNotFoundReason
	return builder
}

func (builder *conditionsBuilder) ApplicationNotFound() *conditionsBuilder {
	builder.reason = ApplicationNotFoundReason
	return builder
}

func (builder *conditionsBuilder) Msg(msg string) *conditionsBuilder {
	builder.message = msg
	return builder
}
