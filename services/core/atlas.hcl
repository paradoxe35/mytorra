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

variable "envfile" {
    type    = string
    default = ".env"
}

locals {
    envfile = {
        for line in split("\n", file(var.envfile)): split("=", line)[0] => regex("=(.*)", line)[0]
        if !startswith(line, "#") && length(split("=", line)) > 1
    }
}


env "bun" {
  src = data.external_schema.bun.url
  dev = docker.postgres.dev.url
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
