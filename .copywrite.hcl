project {
  license = "MPL-2.0"
  copyright_year = 2021
  header_ignore = [
    "*.hcl2spec.go", # generated code specs, since they'll be wiped out until we support adding headers to generated files.
    "**/test-fixtures/**",
    "**/examples/**",
  ]
}
