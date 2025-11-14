package merchant

import (
	"context"

	"github.com/dana-id/dana-go/merchant_management/v1"
	"github.com/riyanathariq/dana-enterprise/package/dana"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetMerchantInfo(ctx context.Context, merchantID string) (*merchant_management.QueryMerchantResourceResponse, error) {
	danaClient := dana.InitData()
	merchantResource, _, err := danaClient.MerchantManagementAPI.QueryMerchantResource(ctx).
		QueryMerchantResourceRequest(merchant_management.QueryMerchantResourceRequest{
			RequestMerchantId: merchantID,
			MerchantResourceInfoList: []string{
				"MERCHANT_DEPOSIT_BALANCE",
				"MERCHANT_AVAILABLE_BALANCE",
				"MERCHANT_TOTAL_BALANCE",
			},
		}).
		Execute()
	if err != nil {
		return nil, err
	}
	return merchantResource, nil
}
