#!/bin/bash

# Налаштування кількості партицій для тем Kafka
# Цей скрипт створює теми Kafka з вказаною кількістю партицій або змінює кількість партицій для існуючих тем

# Конфігурація
KAFKA_BROKER="localhost:9092"
COFFEE_ORDERS_TOPIC="coffee_orders"
PROCESSED_ORDERS_TOPIC="processed_orders"
PARTITIONS=3
REPLICATION_FACTOR=1

# Функція для перевірки наявності теми
topic_exists() {
    kafka-topics.sh --bootstrap-server $KAFKA_BROKER --list | grep -q "^$1$"
    return $?
}

# Функція для створення теми
create_topic() {
    echo "Створення теми $1 з $PARTITIONS партиціями..."
    kafka-topics.sh --bootstrap-server $KAFKA_BROKER --create \
        --topic $1 \
        --partitions $PARTITIONS \
        --replication-factor $REPLICATION_FACTOR
}

# Функція для зміни кількості партицій
alter_partitions() {
    echo "Зміна кількості партицій для теми $1 на $PARTITIONS..."
    kafka-topics.sh --bootstrap-server $KAFKA_BROKER --alter \
        --topic $1 \
        --partitions $PARTITIONS
}

# Перевірка наявності теми coffee_orders
if topic_exists $COFFEE_ORDERS_TOPIC; then
    echo "Тема $COFFEE_ORDERS_TOPIC вже існує."
    
    # Отримання поточної кількості партицій
    CURRENT_PARTITIONS=$(kafka-topics.sh --bootstrap-server $KAFKA_BROKER --describe --topic $COFFEE_ORDERS_TOPIC | grep "PartitionCount" | awk '{print $2}')
    
    if [ "$CURRENT_PARTITIONS" -lt "$PARTITIONS" ]; then
        alter_partitions $COFFEE_ORDERS_TOPIC
    else
        echo "Тема $COFFEE_ORDERS_TOPIC вже має $CURRENT_PARTITIONS партицій, що більше або дорівнює бажаній кількості $PARTITIONS."
    fi
else
    create_topic $COFFEE_ORDERS_TOPIC
fi

# Перевірка наявності теми processed_orders
if topic_exists $PROCESSED_ORDERS_TOPIC; then
    echo "Тема $PROCESSED_ORDERS_TOPIC вже існує."
    
    # Отримання поточної кількості партицій
    CURRENT_PARTITIONS=$(kafka-topics.sh --bootstrap-server $KAFKA_BROKER --describe --topic $PROCESSED_ORDERS_TOPIC | grep "PartitionCount" | awk '{print $2}')
    
    if [ "$CURRENT_PARTITIONS" -lt "$PARTITIONS" ]; then
        alter_partitions $PROCESSED_ORDERS_TOPIC
    else
        echo "Тема $PROCESSED_ORDERS_TOPIC вже має $CURRENT_PARTITIONS партицій, що більше або дорівнює бажаній кількості $PARTITIONS."
    fi
else
    create_topic $PROCESSED_ORDERS_TOPIC
fi

echo "Налаштування тем Kafka завершено."
