// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/": {
            "post": {
                "description": "Create a short URL based on the given URL",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "POST"
                ],
                "summary": "Create new short URL from URL",
                "parameters": [
                    {
                        "description": "URL to shorten",
                        "name": "url",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created"
                    },
                    "400": {
                        "description": "Bad request"
                    },
                    "404": {
                        "description": "URL not found"
                    },
                    "409": {
                        "description": "Conflict"
                    },
                    "500": {
                        "description": "Internal server error"
                    }
                }
            },
            "options": {
                "description": "Verify user",
                "tags": [
                    "AUTH_SERVICE"
                ],
                "summary": "Verify user",
                "parameters": [
                    {
                        "description": "token",
                        "name": "VerifyUser",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "500": {
                        "description": "Internal server error"
                    }
                }
            },
            "patch": {
                "description": "Auth middleware",
                "tags": [
                    "MIDDLEWARE"
                ],
                "summary": "Auth middleware",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Authorization",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "500": {
                        "description": "Internal server error"
                    }
                }
            }
        },
        "/api/shorten": {
            "post": {
                "description": "Create a short URL based on the given JSON payload",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "POST"
                ],
                "summary": "Create new short URL from JSON request",
                "parameters": [
                    {
                        "description": "URL to shorten",
                        "name": "url",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.URL"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created"
                    },
                    "400": {
                        "description": "Bad request"
                    },
                    "404": {
                        "description": "URL not found"
                    },
                    "409": {
                        "description": "Conflict"
                    },
                    "500": {
                        "description": "Internal server error"
                    }
                }
            }
        },
        "/api/shorten/batch": {
            "post": {
                "description": "Create a short URL based on the given URL",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "POST"
                ],
                "summary": "Create new short URL from URL",
                "parameters": [
                    {
                        "description": "URL to shorten",
                        "name": "url",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.MultipleURL"
                            }
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created"
                    },
                    "400": {
                        "description": "Bad request"
                    },
                    "404": {
                        "description": "Not found"
                    },
                    "500": {
                        "description": "Internal server error"
                    }
                }
            }
        },
        "/api/user/urls": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get user URLs",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GET"
                ],
                "summary": "Get user URLs",
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "204": {
                        "description": "No content"
                    },
                    "400": {
                        "description": "Bad request"
                    },
                    "401": {
                        "description": "Unauthorized"
                    },
                    "500": {
                        "description": "Internal server error"
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Delete user URLs",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "DELETE"
                ],
                "summary": "Delete user URLs",
                "parameters": [
                    {
                        "description": "URLs",
                        "name": "urls",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Accepted"
                    },
                    "500": {
                        "description": "Internal server error"
                    }
                }
            },
            "patch": {
                "description": "Check auth middleware",
                "tags": [
                    "MIDDLEWARE"
                ],
                "summary": "Check auth middleware",
                "parameters": [
                    {
                        "type": "string",
                        "description": "user_id",
                        "name": "CheckAuthMiddleware",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "401": {
                        "description": "Unauthorized"
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "description": "Check DB connection",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "GET"
                ],
                "summary": "Check DB connection",
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "500": {
                        "description": "Internal server error"
                    }
                }
            }
        },
        "/pprof/...": {
            "get": {
                "description": "Pprof middleware - work only location",
                "tags": [
                    "MIDDLEWARE"
                ],
                "summary": "Pprof middleware",
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "403": {
                        "description": "Access denied"
                    }
                }
            }
        },
        "/{id}": {
            "get": {
                "description": "Get short URL",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GET"
                ],
                "summary": "Get short URL",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Short URL",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "307": {
                        "description": "Temporary redirect",
                        "headers": {
                            "Location": {
                                "type": "string",
                                "description": "URL новой записи"
                            }
                        }
                    },
                    "404": {
                        "description": "Not found"
                    },
                    "405": {
                        "description": "Method not allowed"
                    },
                    "410": {
                        "description": "Gone"
                    }
                }
            }
        }
    },
    "definitions": {
        "models.MultipleURL": {
            "type": "object",
            "properties": {
                "correlation_id": {
                    "type": "string"
                },
                "original_url": {
                    "type": "string"
                }
            }
        },
        "models.URL": {
            "type": "object",
            "properties": {
                "url": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "URL Shortener API",
	Description:      "API Server for shortener",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
