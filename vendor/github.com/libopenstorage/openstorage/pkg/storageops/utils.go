package storageops

import "time"

// ProviderOpsMaxRetries is the number of retries to use for provider ops
const ProviderOpsMaxRetries = 10

// ProviderOpsRetryInterval is the time to wait before each retry of provider ops
const ProviderOpsRetryInterval = 3 * time.Second

// ProviderOpsTimeout is the default timeout of storage provider ops
const ProviderOpsTimeout = time.Minute

// AddElementToMap adds to the given 'elem' to the 'sets' map with given 'key'
func AddElementToMap(
	sets map[string][]interface{},
	elem interface{},
	key string,
) {
	if s, ok := sets[key]; ok {
		sets[key] = append(s, elem)
	} else {
		sets[key] = make([]interface{}, 0)
		sets[key] = append(sets[key], elem)
	}
}
