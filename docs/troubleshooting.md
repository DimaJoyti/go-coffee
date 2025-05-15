# Troubleshooting Guide

This guide provides solutions to common issues that may arise when using the Coffee Order System.

## Common Issues

### Producer Service

#### Issue: Producer Service Fails to Start

**Symptoms**:
- Error message: `Failed to create Kafka producer: kafka: client has run out of available brokers to talk to`
- Producer service exits immediately

**Possible Causes**:
1. Kafka is not running
2. Kafka is running on a different address than configured
3. Network issues preventing connection to Kafka

**Solutions**:
1. Ensure Kafka is running:
   ```bash
   # Check if Kafka is running in Docker
   docker ps | grep kafka
   ```
2. Check Kafka broker address in configuration:
   ```bash
   # Check environment variable
   echo $KAFKA_BROKERS
   
   # Check configuration file
   cat producer/config.json
   ```
3. Try connecting to Kafka manually:
   ```bash
   # Using kafkacat
   kafkacat -b localhost:9092 -L
   ```

#### Issue: API Returns 500 Internal Server Error

**Symptoms**:
- API returns 500 Internal Server Error
- Error message in logs: `Failed to send message to Kafka`

**Possible Causes**:
1. Kafka is not running
2. Kafka topic does not exist
3. Kafka broker is not accessible

**Solutions**:
1. Ensure Kafka is running (see above)
2. Create the Kafka topic:
   ```bash
   # Using kafka-topics.sh
   kafka-topics.sh --create --topic coffee_orders --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
   ```
3. Check Kafka broker accessibility (see above)

### Consumer Service

#### Issue: Consumer Service Fails to Start

**Symptoms**:
- Error message: `Failed to create Kafka consumer: kafka: client has run out of available brokers to talk to`
- Consumer service exits immediately

**Possible Causes**:
1. Kafka is not running
2. Kafka is running on a different address than configured
3. Network issues preventing connection to Kafka

**Solutions**:
1. Ensure Kafka is running (see above)
2. Check Kafka broker address in configuration (see above)
3. Try connecting to Kafka manually (see above)

#### Issue: Consumer Service Does Not Receive Messages

**Symptoms**:
- Producer service successfully sends messages
- Consumer service does not log any received messages

**Possible Causes**:
1. Consumer is subscribed to a different topic than the producer is publishing to
2. Consumer is not starting from the beginning of the topic
3. Messages are being sent to a different partition

**Solutions**:
1. Check topic configuration in both services:
   ```bash
   # Check producer configuration
   cat producer/config.json
   
   # Check consumer configuration
   cat consumer/config.json
   ```
2. Ensure consumer is starting from the beginning of the topic:
   ```go
   // In consumer/main.go
   consumer, err := worker.ConsumePartition(cfg.Kafka.Topic, 0, sarama.OffsetOldest)
   ```
3. Check Kafka topic partitions:
   ```bash
   # Using kafka-topics.sh
   kafka-topics.sh --describe --topic coffee_orders --bootstrap-server localhost:9092
   ```

## Configuration Issues

### Issue: Configuration File Not Found

**Symptoms**:
- Error message: `Failed to load configuration: open config.json: no such file or directory`

**Possible Causes**:
1. Configuration file does not exist
2. Configuration file is in a different location than expected

**Solutions**:
1. Create the configuration file:
   ```bash
   # For producer
   cp producer/config.json.example producer/config.json
   
   # For consumer
   cp consumer/config.json.example consumer/config.json
   ```
2. Specify the configuration file location:
   ```bash
   export CONFIG_FILE=/path/to/config.json
   ```

### Issue: Invalid Configuration

**Symptoms**:
- Error message: `Failed to load configuration: invalid character...`

**Possible Causes**:
1. Configuration file contains invalid JSON
2. Configuration file is missing required fields

**Solutions**:
1. Validate the JSON:
   ```bash
   # Using jq
   jq . config.json
   ```
2. Check the configuration file against the expected schema (see [Configuration](configuration.md))

## Network Issues

### Issue: Cannot Connect to Kafka

**Symptoms**:
- Error message: `Failed to create Kafka producer: kafka: client has run out of available brokers to talk to`

**Possible Causes**:
1. Kafka is not running
2. Kafka is running on a different address than configured
3. Firewall blocking connection to Kafka

**Solutions**:
1. Ensure Kafka is running (see above)
2. Check Kafka broker address in configuration (see above)
3. Check firewall settings:
   ```bash
   # Check if port is open
   telnet localhost 9092
   ```

## Getting Help

If you're still experiencing issues after trying the solutions in this guide, you can:

1. Check the logs for more detailed error messages:
   ```bash
   # Producer logs
   tail -f producer.log
   
   # Consumer logs
   tail -f consumer.log
   ```
2. Open an issue on the GitHub repository
3. Contact the maintainers

## Next Steps

- [Configuration](configuration.md): Learn about configuration options.
- [API Reference](api-reference.md): Explore the API endpoints.
- [Kafka Integration](kafka-integration.md): Learn about Kafka integration.
