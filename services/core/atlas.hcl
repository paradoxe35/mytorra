data "external_schema" "bun" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-bun",
    "load",
    "--path", "./internal/adapters/postgres",
    "--dialect", "postgres"
  ]
}

env "bun" {
  src = data.external_schema.bun.url
  dev = "docker://postgis/postgis:17-3.4"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
