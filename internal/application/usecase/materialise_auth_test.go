package usecase

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
)

func TestMaterialiseAuthAddsHeaderAuthLikeTheAutoHeader(t *testing.T) {
	bearer := entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: "t"}}
	cases := []struct {
		request entity.EffectiveRequest
		want    []entity.KeyValue
	}{
		{request: entity.EffectiveRequest{Auth: bearer, Headers: []entity.KeyValue{entity.NewKeyValue("A", "1")}}, want: []entity.KeyValue{entity.NewKeyValue("A", "1"), entity.NewKeyValue(entity.HeaderAuthorization, "Bearer t")}},
		{request: entity.EffectiveRequest{Auth: bearer, Headers: []entity.KeyValue{entity.NewKeyValue("authorization", "mine")}}, want: []entity.KeyValue{entity.NewKeyValue("authorization", "mine")}},
		{request: entity.EffectiveRequest{Auth: bearer, DisabledAutoHeaders: []string{"Authorization"}}, want: nil},
		{request: entity.EffectiveRequest{Auth: entity.Auth{Type: authtype.None}}, want: nil},
	}
	for index, tc := range cases {
		if got := MaterialiseAuth(tc.request).Headers; !reflect.DeepEqual(got, tc.want) {
			t.Errorf("case %d: got %+v, want %+v", index, got, tc.want)
		}
	}
}
