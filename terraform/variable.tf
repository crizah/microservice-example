variable "pk" {
    type = string
    default = "username"
    description = "partition key for tables"
  
}

variable "region"{
    type = string
    default = "eu-north-1"
    description = "region of provider"
}

variable "u" {
    type = string
    default = "Users"

}

variable "p" {
    type = string
    default = "Pass"
    
 
}

variable "et" {
    type = string
    default = "Email-token"
  
}

variable "prt" {
    type= string
    default = "Password-reset-token" 
  
}