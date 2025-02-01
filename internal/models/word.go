package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Word struct {
    ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Word          string            `bson:"word" json:"word"`
    Definitions   []Definition      `bson:"definitions" json:"definitions"`
    Length        int               `bson:"length" json:"length"`
    Pronunciation string            `bson:"pronunciation,omitempty" json:"pronunciation,omitempty"`
    Etymology     string            `bson:"etymology,omitempty" json:"etymology,omitempty"`
    Synonyms      []string          `bson:"synonyms,omitempty" json:"synonyms,omitempty"`
    Antonyms      []string          `bson:"antonyms,omitempty" json:"antonyms,omitempty"`
    Examples      []string          `bson:"examples,omitempty" json:"examples,omitempty"`
    PartOfSpeech  string            `bson:"partOfSpeech,omitempty" json:"partOfSpeech,omitempty"`
    Usage         string            `bson:"usage,omitempty" json:"usage,omitempty"`
    Frequency     float64           `bson:"frequency,omitempty" json:"frequency,omitempty"`
    CreatedAt     primitive.DateTime `bson:"createdAt" json:"createdAt"`
    UpdatedAt     primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

type Definition struct {
    Meaning     string   `bson:"meaning" json:"meaning"`
    PartOfSpeech string  `bson:"partOfSpeech,omitempty" json:"partOfSpeech,omitempty"`
    Examples    []string `bson:"examples,omitempty" json:"examples,omitempty"`
}

type SearchOptions struct {
    StartsWith string
    EndsWith   string
    Contains   string
    Length     int
    Limit      int
    Skip       int
    SortBy     string
    SortOrder  int
}

type SearchResponse struct {
    Words      []Word `json:"words"`
    TotalCount int64  `json:"totalCount"`
    Page       int    `json:"page"`
    PageSize   int    `json:"pageSize"`
}

type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string     `json:"error,omitempty"`
    Meta    *MetaData  `json:"meta,omitempty"`
}

type MetaData struct {
    TotalCount  int64 `json:"totalCount,omitempty"`
    CurrentPage int   `json:"currentPage,omitempty"`
    PageSize    int   `json:"pageSize,omitempty"`
    TotalPages  int   `json:"totalPages,omitempty"`
}