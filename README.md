# learn-pub-sub-starter (Peril)

This is the starter code used in Boot.dev's [Learn Pub/Sub](https://learn.boot.dev/learn-pub-sub) course.


## RabbitMQ setup:

1. Go to the RabbitMQ management UI at http://localhost:15672
2. navigate to the "Exchanges" tab
3. Create a new exchange called `peril_direct` with the type `direct`
4. Create a new exchange with type `topic`, and name it `peril_topic`.

See https://www.rabbitmq.com/tutorials/amqp-concepts#exchange-topic
