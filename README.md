Bundle spliter
====

Usage

    /path/to/bundle-spliter
        -url "amqp://go1:go1@127.0.0.1:5672/"
        -kind "topic"
        -exchange "events"
        -routing-keys "ro.create,ro.update,ro.delete"
        -queue-name "ro-bundle-spliter"
        -consumer-name "ro-bundle-spliter"

Bundle spliter is a tiny application that listens into some routing-keys split it into smaller bundle queues. Example:

- Input:
    - routing key: `ro.create`
    - body: `{ "type": "has-ro", "foo": "bar" }`
- The same message will be published to same channel with:
    - routing key: `ro.create.has-ro`
    - body: `{ "type": "has-ro", "foo": "bar" }`
