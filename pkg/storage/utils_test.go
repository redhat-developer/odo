package storage

func generateStorage(storage Storage, status StorageStatus, containerName string) Storage {
	storage.Status = status
	storage.Spec.ContainerName = containerName
	return storage
}
