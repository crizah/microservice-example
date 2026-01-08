
# create dynamodb tables



resource "aws_dynamodb_table" "table1" {

  name           = var.u
  billing_mode = "PAY_PER_REQUEST"
  hash_key       = var.pk

  attribute {
    name = var.pk
    type = "S"
  }

  attribute {
    name = "email"
    type = "S"
  }

  global_secondary_index {
    name               = "EmailIndex"
    hash_key           = "email"
    projection_type    = "ALL"
   
  }

}



resource "aws_dynamodb_table" "table2" {

  name           = var.p
  billing_mode = "PAY_PER_REQUEST"
  hash_key       = var.pk

  attribute {
    name = var.pk
    type = "S"
  }
}


resource "aws_dynamodb_table" "table3" {

  name           = var.et
  billing_mode = "PAY_PER_REQUEST"
  hash_key       = var.pk

  attribute {
    name = var.pk
    type = "S"
  }

  attribute {
    name = "token"
    type = "S"
  }

  global_secondary_index {
    name               = "TokenIndex"
    hash_key           = "token"
    projection_type    = "ALL"
   
  }

}

resource "aws_dynamodb_table" "table4" {

  name           = var.prt
  billing_mode = "PAY_PER_REQUEST"
  hash_key       = var.pk

  attribute {
    name = var.pk
    type = "S"
  }
}

