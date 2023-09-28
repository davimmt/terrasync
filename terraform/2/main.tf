output "out" {
  value = 100 + basename(abspath(path.module))
}
