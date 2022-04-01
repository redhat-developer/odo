package kclient

import (
	"errors"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestClient_TryWithBlockOwnerDeletion(t *testing.T) {
	type args struct {
		ownerReference metav1.OwnerReference
		exec           func() func(ownerReference metav1.OwnerReference) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "first call is ok",
			args: args{
				ownerReference: metav1.OwnerReference{},
				exec: func() func(ownerReference metav1.OwnerReference) error {
					calls := 0
					return func(ownerReference metav1.OwnerReference) error {
						calls++
						if calls > 1 {
							t.Errorf("only one call should happen")
						}
						return nil
					}
				},
			},
			wantErr: false,
		},
		{
			name: "first call fails with non forbidden error",
			args: args{
				ownerReference: metav1.OwnerReference{},
				exec: func() func(ownerReference metav1.OwnerReference) error {
					calls := 0
					return func(ownerReference metav1.OwnerReference) error {
						calls++
						if calls > 1 {
							t.Errorf("only one call should happen")
						}
						return errors.New("an error call 1")
					}
				},
			},
			wantErr: true,
		},
		{
			name: "first call fails with forbidden error, second call is ok",
			args: args{
				ownerReference: metav1.OwnerReference{},
				exec: func() func(ownerReference metav1.OwnerReference) error {
					calls := 0
					return func(ownerReference metav1.OwnerReference) error {
						calls++
						switch calls {
						case 1:
							return apierrors.NewForbidden(schema.GroupResource{}, "aname", errors.New("an error"))
						case 2:
							return nil
						default:
							t.Errorf("only two calls should happen")
							return nil
						}
					}
				},
			},
			wantErr: false,
		},
		{
			name: "first call fails with forbidden error, second call fails with error",
			args: args{
				ownerReference: metav1.OwnerReference{},
				exec: func() func(ownerReference metav1.OwnerReference) error {
					calls := 0
					return func(ownerReference metav1.OwnerReference) error {
						calls++
						switch calls {
						case 1:
							return apierrors.NewForbidden(schema.GroupResource{}, "aname", errors.New("an error"))
						case 2:
							return errors.New("an error")
						default:
							t.Errorf("only two calls should happen")
							return nil
						}
					}
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{}
			if err := c.TryWithBlockOwnerDeletion(tt.args.ownerReference, tt.args.exec()); (err != nil) != tt.wantErr {
				t.Errorf("Client.TryWithBlockOwnerDeletion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
