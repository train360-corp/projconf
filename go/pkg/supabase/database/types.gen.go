package database

type PublicProjectsSelect struct {
  Display string `json:"display"`
  Id      string `json:"id"`
}

type PublicProjectsInsert struct {
  Display *string `json:"display"`
  Id      *string `json:"id"`
}

type PublicProjectsUpdate struct {
  Display *string `json:"display"`
  Id      *string `json:"id"`
}

type PublicVariablesSelect struct {
  Description   string      `json:"description"`
  GeneratorData interface{} `json:"generator_data"`
  GeneratorType string      `json:"generator_type"`
  Id            string      `json:"id"`
  Key           string      `json:"key"`
  ProjectId     string      `json:"project_id"`
}

type PublicVariablesInsert struct {
  Description   *string     `json:"description"`
  GeneratorData interface{} `json:"generator_data"`
  GeneratorType string      `json:"generator_type"`
  Id            *string     `json:"id"`
  Key           string      `json:"key"`
  ProjectId     string      `json:"project_id"`
}

type PublicVariablesUpdate struct {
  Description   *string     `json:"description"`
  GeneratorData interface{} `json:"generator_data"`
  GeneratorType *string     `json:"generator_type"`
  Id            *string     `json:"id"`
  Key           *string     `json:"key"`
  ProjectId     *string     `json:"project_id"`
}

type PublicClientsSelect struct {
  CreatedAt     string `json:"created_at"`
  Display       string `json:"display"`
  EnvironmentId string `json:"environment_id"`
  Id            string `json:"id"`
}

type PublicClientsInsert struct {
  CreatedAt     *string `json:"created_at"`
  Display       *string `json:"display"`
  EnvironmentId string  `json:"environment_id"`
  Id            *string `json:"id"`
}

type PublicClientsUpdate struct {
  CreatedAt     *string `json:"created_at"`
  Display       *string `json:"display"`
  EnvironmentId *string `json:"environment_id"`
  Id            *string `json:"id"`
}

type PublicClientsSecretsSelect struct {
  ClientId  string `json:"client_id"`
  CreatedAt string `json:"created_at"`
  Hash      string `json:"hash"`
  Id        string `json:"id"`
}

type PublicClientsSecretsInsert struct {
  ClientId  string  `json:"client_id"`
  CreatedAt *string `json:"created_at"`
  Hash      string  `json:"hash"`
  Id        *string `json:"id"`
}

type PublicClientsSecretsUpdate struct {
  ClientId  *string `json:"client_id"`
  CreatedAt *string `json:"created_at"`
  Hash      *string `json:"hash"`
  Id        *string `json:"id"`
}

type PublicEnvironmentsSelect struct {
  CreatedAt string `json:"created_at"`
  Display   string `json:"display"`
  Id        string `json:"id"`
  ProjectId string `json:"project_id"`
}

type PublicEnvironmentsInsert struct {
  CreatedAt *string `json:"created_at"`
  Display   *string `json:"display"`
  Id        *string `json:"id"`
  ProjectId string  `json:"project_id"`
}

type PublicEnvironmentsUpdate struct {
  CreatedAt *string `json:"created_at"`
  Display   *string `json:"display"`
  Id        *string `json:"id"`
  ProjectId *string `json:"project_id"`
}

type PublicSecretsSelect struct {
  CreatedAt     string `json:"created_at"`
  EnvironmentId string `json:"environment_id"`
  Id            string `json:"id"`
  Value         string `json:"value"`
  VariableId    string `json:"variable_id"`
}

type PublicSecretsInsert struct {
  CreatedAt     *string `json:"created_at"`
  EnvironmentId string  `json:"environment_id"`
  Id            *string `json:"id"`
  Value         *string `json:"value"`
  VariableId    string  `json:"variable_id"`
}

type PublicSecretsUpdate struct {
  CreatedAt     *string `json:"created_at"`
  EnvironmentId *string `json:"environment_id"`
  Id            *string `json:"id"`
  Value         *string `json:"value"`
  VariableId    *string `json:"variable_id"`
}
