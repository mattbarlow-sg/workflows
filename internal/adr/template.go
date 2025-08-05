package adr

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ADR struct {
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	Status          string           `json:"status"`
	Date            string           `json:"date"`
	TechnicalStory  *TechnicalStory  `json:"technicalStory,omitempty"`
	Context         Context          `json:"context"`
	DecisionDrivers []DecisionDriver `json:"decisionDrivers,omitempty"`
	Options         []Option         `json:"options"`
	Decision        Decision         `json:"decision"`
	Validation      *Validation      `json:"validation,omitempty"`
	Consequences    Consequences     `json:"consequences"`
	Compliance      *Compliance      `json:"compliance,omitempty"`
	Notes           *string          `json:"notes,omitempty"`
	Stakeholders    *Stakeholders    `json:"stakeholders,omitempty"`
	AIMetadata      *AIMetadata      `json:"aiMetadata,omitempty"`
	Links           []Link           `json:"links,omitempty"`
	WorkSessionID   string           `json:"workSessionId,omitempty"`
}

type TechnicalStory struct {
	ID                 string   `json:"id,omitempty"`
	Title              string   `json:"title,omitempty"`
	Description        string   `json:"description,omitempty"`
	AcceptanceCriteria []string `json:"acceptanceCriteria,omitempty"`
	Link               string   `json:"link,omitempty"`
}

type Context struct {
	Problem     string   `json:"problem"`
	Background  string   `json:"background"`
	Constraints []string `json:"constraints,omitempty"`
	Assumptions []string `json:"assumptions,omitempty"`
}

type DecisionDriver struct {
	Driver      string  `json:"driver"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description,omitempty"`
}

type Option struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Pros        []string `json:"pros,omitempty"`
	Cons        []string `json:"cons,omitempty"`
	Score       *float64 `json:"score,omitempty"`
}

type Decision struct {
	ChosenOption   string `json:"chosenOption"`
	Rationale      string `json:"rationale"`
	Implementation string `json:"implementation,omitempty"`
}

type Validation struct {
	Method         string   `json:"method,omitempty"`
	SuccessCriteria []string `json:"successCriteria,omitempty"`
	Metrics        []Metric `json:"metrics,omitempty"`
}

type Metric struct {
	Metric    string `json:"metric"`
	Target    string `json:"target"`
	Timeframe string `json:"timeframe,omitempty"`
}

type Consequences struct {
	Positive []string `json:"positive,omitempty"`
	Negative []string `json:"negative,omitempty"`
	Neutral  []string `json:"neutral,omitempty"`
}

type Compliance struct {
	Standards   []string `json:"standards,omitempty"`
	Regulations []string `json:"regulations,omitempty"`
}

type Stakeholders struct {
	Deciders   []string `json:"deciders,omitempty"`
	Consulted  []string `json:"consulted,omitempty"`
	Informed   []string `json:"informed,omitempty"`
}

type AIMetadata struct {
	Tags           []string      `json:"tags,omitempty"`
	ImpactScores   *ImpactScores `json:"impactScores,omitempty"`
	Dependencies   []Dependency  `json:"dependencies,omitempty"`
	EstimatedCost  *Cost         `json:"estimatedCost,omitempty"`
	Keywords       []string      `json:"keywords,omitempty"`
}

type ImpactScores struct {
	Technical float64 `json:"technical,omitempty"`
	Business  float64 `json:"business,omitempty"`
	Risk      float64 `json:"risk,omitempty"`
}

type Dependency struct {
	ADRID        string `json:"adrId"`
	Relationship string `json:"relationship"`
}

type Cost struct {
	Development string `json:"development,omitempty"`
	Maintenance string `json:"maintenance,omitempty"`
}

type Link struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type,omitempty"`
}

func GenerateTemplate(title string, deciders []string) *ADR {
	id := fmt.Sprintf("ADR-%04d", time.Now().Unix()%10000)
	date := time.Now().Format("2006-01-02")

	stakeholders := &Stakeholders{
		Deciders: deciders,
	}

	return &ADR{
		ID:     id,
		Title:  title,
		Status: "draft",
		Date:   date,
		Context: Context{
			Problem:    "",
			Background: "",
		},
		Options: []Option{},
		Decision: Decision{
			ChosenOption: "",
			Rationale:    "",
		},
		Consequences: Consequences{
			Positive: []string{},
			Negative: []string{},
		},
		Stakeholders: stakeholders,
	}
}

func (a *ADR) ToJSON() (string, error) {
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func FromJSON(data []byte) (*ADR, error) {
	var adr ADR
	err := json.Unmarshal(data, &adr)
	if err != nil {
		return nil, err
	}
	return &adr, nil
}

func GenerateFilename(adr *ADR) string {
	timestamp := time.Now().Format("20060102")
	safeTitle := sanitizeFilename(adr.Title)
	return fmt.Sprintf("%s-%s-%s.json", adr.ID, timestamp, safeTitle)
}

func sanitizeFilename(s string) string {
	result := ""
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result += string(r)
		} else if r == ' ' {
			result += "-"
		}
	}
	return strings.ToLower(result)
}