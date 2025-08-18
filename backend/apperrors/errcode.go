package apperrors

type ErrCode string

const (
	// Unknown: 予期せぬエラー
	Unknown ErrCode = "U000"

	// InsertDataFailed: データ挿入失敗
	InsertDataFailed ErrCode = "S001"
	// GetDataFailed: データ取得失敗
	GetDataFailed ErrCode = "S002"
	// NoData: データが存在しない
	NoData ErrCode = "S003"
	// UpdateDataFailed: データ更新失敗
	UpdateDataFailed ErrCode = "S004"
	// DeleteDataFailed: データ削除失敗
	DeleteDataFailed ErrCode = "S005"

	// ReqBodyDecodeFailed: リクエストボディのデコード失敗
	ReqBodyDecodeFailed ErrCode = "R001"
	// BadParam: URLパラメータなどが不正
	BadParam ErrCode = "R002"
	// ValidationFailed: バリデーション失敗
	ValidationFailed ErrCode = "R003"

	// Unauthorized: 認証エラー（トークンがない、不正など）
	Unauthorized ErrCode = "A001"
	// Forbidden: 認可エラー（権限がない）
	Forbidden ErrCode = "A002"

	// Conflict: 状態の競合（メールアドレスの重複、ステータスの不整合など）
	Conflict ErrCode = "C001"
)