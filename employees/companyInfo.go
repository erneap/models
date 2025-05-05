package employees

type CompanyInfo struct {
	Company     string `json:"company"`
	EmployeeID  string `json:"employeeid"`
	AlternateID string `json:"alternateid"`
	JobTitle    string `json:"jobtitle"`
	Rank        string `json:"rank"`
	CostCenter  string `json:"costcenter"`
	Division    string `json:"division"`
}
