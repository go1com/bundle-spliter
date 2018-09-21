Bundle splitter
====

Usage

    /path/to/bundle-splitter
        -url "amqp://go1:go1@127.0.0.1:5672/"
        -kind "topic"
        -exchange "events"
        -routing-keys "ro.create,ro.update,ro.delete"
        -queue-name "ro-bundle-splitter"
        -consumer-name "ro-bundle-splitter"

Bundle splitter is a tiny application that listens into some routing-keys split it into smaller bundle queues. Example:

- Input:
    - routing key: `ro.create`
    - body: `{ "type": "has-ro", "foo": "bar" }`
- The same message will be published to same channel with:
    - routing key: `ro.create.has-ro`
    - body: `{ "type": "has-ro", "foo": "bar" }`
