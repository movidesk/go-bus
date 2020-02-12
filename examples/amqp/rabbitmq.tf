provider "rabbitmq" {
    endpoint = "http://localhost:15672"
    username = "guest"
    password = "guest"
}

resource "rabbitmq_exchange" "events" {
  name  = "events"
  vhost = "/"

  settings {
    type        = "fanout"
    durable     = true 
    auto_delete = false
  }
}

resource "rabbitmq_queue" "event_queue_a" {
  name  = "event_queue_a"
  vhost = "/"

  settings {
    durable     = true 
    auto_delete = false
  }

  depends_on = [
      "rabbitmq_exchange.events"
  ]
}

resource "rabbitmq_queue" "event_queue_b" {
  name  = "event_queue_b"
  vhost = "/"

  settings {
    durable     = true 
    auto_delete = false
  }

  depends_on = [
      "rabbitmq_exchange.events"
  ]
}

resource "rabbitmq_binding" "bind_event_queue_a" {
  source           = "events"
  vhost            = "/"
  destination      = "event_queue_a"
  destination_type = "queue"
  routing_key      = ""

  depends_on = [
      "rabbitmq_exchange.events",
      "rabbitmq_queue.event_queue_a"
  ]
}

resource "rabbitmq_binding" "bind_event_queue_b" {
  source           = "events"
  vhost            = "/"
  destination      = "event_queue_b"
  destination_type = "queue"
  routing_key      = ""

  depends_on = [
      "rabbitmq_exchange.events",
      "rabbitmq_queue.event_queue_b"
  ]
}

resource "rabbitmq_exchange" "commands" {
  name  = "commands"
  vhost = "/"

  settings {
    type        = "topic"
    durable     = true 
    auto_delete = false
  }
}

resource "rabbitmq_queue" "command_queue" {
  name  = "command_queue"
  vhost = "/"

  settings {
    durable     = true 
    auto_delete = false
  }

  depends_on = [
      "rabbitmq_exchange.commands"
  ]
}

resource "rabbitmq_binding" "bind_command_queue" {
  source           = "commands"
  vhost            = "/"
  destination      = "command_queue"
  destination_type = "queue"
  routing_key      = ""

  depends_on = [
      "rabbitmq_exchange.commands",
      "rabbitmq_queue.command_queue"
  ]
}