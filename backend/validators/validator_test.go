package validators

import (
	"testing"
)

// テスト用の構造体
type TestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"min=0,max=150"`
}

type ComplexTestStruct struct {
	Name     string `validate:"required,min=2,max=50"`
	Email    string `validate:"required,email"`
	Age      int    `validate:"min=0,max=150"`
	Password string `validate:"required,min=8"`
	Phone    string `validate:"omitempty,len=11"`
	Website  string `validate:"omitempty,url"`
}

func TestValidator(t *testing.T) {
	t.Run("TestStruct", func(t *testing.T) {
		tests := []struct {
			name      string
			testData  TestStruct
			wantError bool
		}{
			{
				name: "正常系: 全てのフィールドが有効",
				testData: TestStruct{
					Name:  "Test User",
					Email: "test@example.com",
					Age:   25,
				},
				wantError: false,
			},
			{
				name: "異常系: Nameが空",
				testData: TestStruct{
					Name:  "", // required field is empty
					Email: "test@example.com",
					Age:   25,
				},
				wantError: true,
			},
			{
				name: "異常系: Emailが不正",
				testData: TestStruct{
					Name:  "Test User",
					Email: "invalid-email",
					Age:   25,
				},
				wantError: true,
			},
			{
				name: "異常系: 年齢が負の値",
				testData: TestStruct{
					Name:  "Test User",
					Email: "test@example.com",
					Age:   -1, // below minimum
				},
				wantError: true,
			},
			{
				name: "異常系: 年齢が上限を超過",
				testData: TestStruct{
					Name:  "Test User",
					Email: "test@example.com",
					Age:   151, // above maximum
				},
				wantError: true,
			},
			{
				name: "異常系: 複数フィールドが不正",
				testData: TestStruct{
					Name:  "", // required field is empty
					Email: "invalid-email",
					Age:   -1, // below minimum
				},
				wantError: true,
			},
		}

		validator := NewValidator[TestStruct]()

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.Validate(tt.testData)
				if (err != nil) != tt.wantError {
					t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
				}
			})
		}
	})

	t.Run("TestStruct_Pointer", func(t *testing.T) {
		tests := []struct {
			name      string
			testData  *TestStruct
			wantError bool
		}{
			{
				name: "正常系: ポインタ型で全てのフィールドが有効",
				testData: &TestStruct{
					Name:  "Test User",
					Email: "test@example.com",
					Age:   25,
				},
				wantError: false,
			},
			{
				name: "異常系: ポインタ型でバリデーションエラー",
				testData: &TestStruct{
					Name:  "",
					Email: "invalid-email",
					Age:   -1,
				},
				wantError: true,
			},
		}

		validator := NewValidator[*TestStruct]()

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.Validate(tt.testData)
				if (err != nil) != tt.wantError {
					t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
				}
			})
		}
	})

	t.Run("ComplexStruct", func(t *testing.T) {
		tests := []struct {
			name      string
			testData  ComplexTestStruct
			wantError bool
		}{
			{
				name: "正常系: 全てのフィールドが有効",
				testData: ComplexTestStruct{
					Name:     "Test User",
					Email:    "test@example.com",
					Age:      25,
					Password: "password123",
					Phone:    "09012345678",
					Website:  "https://example.com",
				},
				wantError: false,
			},
			{
				name: "正常系: オプショナルフィールドが空",
				testData: ComplexTestStruct{
					Name:     "Test User",
					Email:    "test@example.com",
					Age:      25,
					Password: "password123",
					Phone:    "", // omitempty
					Website:  "", // omitempty
				},
				wantError: false,
			},
			{
				name: "異常系: 名前が短すぎる",
				testData: ComplexTestStruct{
					Name:     "X", // too short (min=2)
					Email:    "test@example.com",
					Age:      25,
					Password: "password123",
					Phone:    "",
					Website:  "",
				},
				wantError: true,
			},
			{
				name: "異常系: パスワードが短すぎる",
				testData: ComplexTestStruct{
					Name:     "Test User",
					Email:    "test@example.com",
					Age:      25,
					Password: "123", // too short (min=8)
					Phone:    "",
					Website:  "",
				},
				wantError: true,
			},
			{
				name: "異常系: 電話番号の長さが不正",
				testData: ComplexTestStruct{
					Name:     "Test User",
					Email:    "test@example.com",
					Age:      25,
					Password: "password123",
					Phone:    "123", // wrong length (should be 11)
					Website:  "",
				},
				wantError: true,
			},
			{
				name: "異常系: WebサイトURLが不正",
				testData: ComplexTestStruct{
					Name:     "Test User",
					Email:    "test@example.com",
					Age:      25,
					Password: "password123",
					Phone:    "",
					Website:  "not-a-url", // invalid URL
				},
				wantError: true,
			},
			{
				name: "異常系: 複数フィールドが不正",
				testData: ComplexTestStruct{
					Name:     "X", // too short
					Email:    "invalid-email",
					Age:      -1,    // below minimum
					Password: "123", // too short
					Phone:    "123", // wrong length
					Website:  "not-a-url",
				},
				wantError: true,
			},
		}

		validator := NewValidator[ComplexTestStruct]()

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.Validate(tt.testData)
				if (err != nil) != tt.wantError {
					t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
				}
			})
		}
	})
}
