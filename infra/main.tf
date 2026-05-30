
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = "supplychain-prod"
  cluster_version = "1.31"           
  vpc_id          = module.vpc.vpc_id
  subnet_ids      = module.vpc.private_subnets

#   eks_managed_node_groups = { }
}

resource "helm_release" "argocd" {
  name       = "argocd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  namespace  = "argocd"
  create_namespace = true

  depends_on = [module.eks]
}