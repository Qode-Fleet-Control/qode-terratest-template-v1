variable "prefix" {
  description = "Prefix of the generated name."
  type        = string
  default     = "fleet"
}

variable "pet_length" {
  description = "Number of words in the generated name."
  type        = number
  default     = 2

  validation {
    condition     = var.pet_length >= 1 && var.pet_length <= 5
    error_message = "pet_length must be between 1 and 5."
  }
}

variable "output_dir" {
  description = "Directory the greeting file is written to."
  type        = string
  default     = "out"
}
