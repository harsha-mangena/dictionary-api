package models

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Definition struct {
    PartOfSpeech string `bson:"part_of_speech" json:"part_of_speech"`
    Definition   string `bson:"definition" json:"definition"`
}


type Word struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Word        string            `bson:"word" json:"word"`
    Definitions []Definition      `bson:"definitions" json:"definitions"`
    Length      int               `bson:"length" json:"length"`
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