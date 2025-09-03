data "external_schema" "bun" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-bun",
    "load",
    "--path", "./internal/adapters/postgres/models",
    "--dialect", "postgres"
  ]
}


env "bun" {
  src = data.external_schema.bun.url
  dev = "docker://postgis/17-3.5-alpine/dev"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
