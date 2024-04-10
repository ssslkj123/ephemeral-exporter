package types

type StatsSummaryForPodEphemeralStorage struct {
	Pods []struct {
		PodRef struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"podRef"`
		EphemeralStorage struct {
			UsedBytes int64 `json:"usedBytes"`
		} `json:"ephemeral-storage"`
	} `json:"pods"`
}

type PodEphemeralStorage struct {
	Name        string
	Namespace   string
	UsageBytes  int64
	LimitBytes  int64
	UsedPercent float64
}
