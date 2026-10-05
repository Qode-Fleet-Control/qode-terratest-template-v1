resource "random_pet" "this" {
  length    = var.pet_length
  prefix    = var.prefix
  separator = "-"
}

resource "local_file" "greeting" {
  filename        = "${var.output_dir}/${random_pet.this.id}.txt"
  content         = "Hello, ${random_pet.this.id}!\n"
  file_permission = "0644"
}
