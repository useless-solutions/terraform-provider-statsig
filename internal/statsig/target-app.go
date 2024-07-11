package statsig

type CreateTargetAppAPIRequest struct {
  Name           string   `json:"name"`
  Description    string   `json:"description"`
  Gates          []string `json:"gates"`
  DynamicConfigs []string `json:"dynamicConfigs"`
  Experiments    []string `json:"experiments"`
}

type CreateTargetAppAPIResponse struct {
  Message string      `json:"message"`
  Data    []TargetApp `json:"data"`
}

type TargetApp struct {
  Name string `json:"name"`
  ID   string `json:"id"`
}
