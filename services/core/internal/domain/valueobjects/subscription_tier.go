package valueobjects

type SubscriptionTier string

const (
	TierBasic        SubscriptionTier = "basic"
	TierProfessional SubscriptionTier = "professional"
	TierPremium      SubscriptionTier = "premium"
	TierEnterprise   SubscriptionTier = "enterprise"
)

func (t SubscriptionTier) String() string {
	return string(t)
}

func (t SubscriptionTier) IsValid() bool {
	switch t {
	case TierBasic, TierProfessional, TierPremium, TierEnterprise:
		return true
	default:
		return false
	}
}

func (t SubscriptionTier) GetPriority() int {
	switch t {
	case TierBasic:
		return 1
	case TierProfessional:
		return 2
	case TierPremium:
		return 3
	case TierEnterprise:
		return 4
	default:
		return 0
	}
}
