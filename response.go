package main

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type MusicListElement struct {
	FileName     string `json:"fileName"`
	FileSize     int64  `json:"fileSize"`
	ModifiedTime string `json:"modifiedTime"`
}

type FileListData struct {
	MusicList     []MusicListElement `json:"musicList"`
	SubFolderList []string           `json:"subFolderList"`
}

type FileListResult struct {
	Type string       `json:"type"`
	Data FileListData `json:"data"`
}

type FileListResponse struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Result  FileListResult `json:"result"`
}

func NewFileListResponse(musicList []MusicListElement, subFloderList []string) FileListResponse {
	return FileListResponse{
		Status:  200,
		Message: "OK",
		Result: FileListResult{
			Type: "fileList",
			Data: FileListData{
				MusicList:     musicList,
				SubFolderList: subFloderList,
			},
		},
	}
}

func NewErrorResponse(status int, message string) ErrorResponse {
	return ErrorResponse{
		Status:  status,
		Message: message,
	}
}
