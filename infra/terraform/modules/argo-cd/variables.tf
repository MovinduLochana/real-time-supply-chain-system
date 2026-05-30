variable "namespace" {
  type = string
}

variable "argocd_version" {
  type    = string
  default = "6.0.0"
}

variable "argocd_domain" {
  type    = string
  default = "argocd.localhost"
}

variable "github_repo_url" {
  type = string
}
