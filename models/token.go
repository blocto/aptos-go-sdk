package models

type CollectionMutabilityConfig struct {
	Description bool
	URI         bool
	Maximum     bool
}

type CollectionData struct {
	Name         string
	Description  string
	URI          string
	Maximum      Uint64
	Supply       Uint64
	MutateConfig CollectionMutabilityConfig `json:"mutability_config"`
}

type TokenMutabilityConfig struct {
	Maximum     bool
	URI         bool
	Description bool
	Royalty     bool
	Properties  bool
}

type TokenData struct {
	Collection   string                `json:"collection"`
	Description  string                `json:"description"`
	Name         string                `json:"name"`
	Maximum      Uint64                `json:"maximum"`
	Supply       Uint64                `json:"supply"`
	URI          string                `json:"uri"`
	MutateConfig TokenMutabilityConfig `json:"mutability_config"`
}

type TokenDataID struct {
	Creator    string `json:"creator"`
	Collection string `json:"collection"`
	Name       string `json:"name"`
}

type TokenID struct {
	TokenDataID     `json:"token_data_id" mapstructure:"token_data_id"`
	PropertyVersion string `json:"property_version" mapstructure:"property_version"`
}

type Token struct {
	ID              TokenID     `json:"id"`
	Amount          Uint64      `json:"amount"`
	TokenProperties PropertyMap `json:"token_properties"`
}

type PropertyMap struct {
	Map SimpleMap `json:"map"`
}

type SimpleMap struct {
	Data []struct {
		Key   string        `json:"key"`
		Value PropertyValue `json:"value"`
	} `json:"data"`
}

type PropertyValue struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
