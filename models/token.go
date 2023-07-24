package models

type CollectionMutabilityConfig struct {
	Description bool `json:"description"`
	URI         bool `json:"uri"`
	Maximum     bool `jons:"maximum"`
}

type CollectionData struct {
	Name         string                     `json:"name"`
	Description  string                     `json:"description"`
	URI          string                     `json:"uri"`
	Maximum      Uint64                     `json:"maximum"`
	Supply       Uint64                     `json:"supply"`
	MutateConfig CollectionMutabilityConfig `json:"mutability_config"`
}

type TokenMutabilityConfig struct {
	Maximum     bool `json:"maximum"`
	URI         bool `json:"uri"`
	Description bool `json:"description"`
	Royalty     bool `json:"royalty"`
	Properties  bool `json:"properties"`
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
	Hash       string `json:"hash"`
	Creator    string `json:"creator"`
	Collection string `json:"collection"`
	Name       string `json:"name"`
}

type TokenID struct {
	TokenDataID     `json:"token_data_id"`
	PropertyVersion Uint64 `json:"property_version"`
}

type Token struct {
	ID              TokenID           `json:"id"`
	Amount          Uint64            `json:"amount"`
	TokenProperties PropertyMap       `json:"token_properties"`
	JSONProperties  map[string]string `json:"-"`
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

type TokenV2 struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	URI               string  `json:"uri"`
	Standard          string  `json:"standard"`
	OwnerAddress      string  `json:"owner_address"`
	Amount            Uint64  `json:"amount"`
	CollectionName    string  `json:"collection_name"`
	CreatorAddress    string  `json:"creator_address"`
	Maximum           *Uint64 `json:"maximum"`
	Supply            Uint64  `json:"supply"`
	PropertyVersionV1 Uint64  `json:"property_version_v1"`
	IsSoulboundV2     bool    `json:"is_soulbound_v2"`
}
