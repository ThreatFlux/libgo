package api

// StorageHandlers holds all storage-related handlers
type StorageHandlers struct {
	// Pool handlers
	ListPools  Handler
	CreatePool Handler
	GetPool    Handler
	DeletePool Handler
	StartPool  Handler
	StopPool   Handler

	// Volume handlers
	ListVolumes  Handler
	CreateVolume Handler
	DeleteVolume Handler
	UploadVolume Handler
}
