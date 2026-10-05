output "name" {
  description = "Generated name."
  value       = random_pet.this.id
}

output "file_path" {
  description = "Path of the greeting file."
  value       = local_file.greeting.filename
}
