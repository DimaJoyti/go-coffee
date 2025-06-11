@echo off
setlocal enabledelayedexpansion

REM Налаштування кількості партицій для тем Kafka
REM Цей скрипт створює теми Kafka з вказаною кількістю партицій або змінює кількість партицій для існуючих тем

REM Конфігурація
set KAFKA_BROKER=localhost:9092
set COFFEE_ORDERS_TOPIC=coffee_orders
set PROCESSED_ORDERS_TOPIC=processed_orders
set PARTITIONS=3
set REPLICATION_FACTOR=1

REM Перевірка наявності теми coffee_orders
echo Перевірка наявності теми %COFFEE_ORDERS_TOPIC%...
kafka-topics.bat --bootstrap-server %KAFKA_BROKER% --list | findstr /B /C:"%COFFEE_ORDERS_TOPIC%" > nul
if %ERRORLEVEL% == 0 (
    echo Тема %COFFEE_ORDERS_TOPIC% вже існує.
    
    REM Отримання поточної кількості партицій
    for /f "tokens=2" %%i in ('kafka-topics.bat --bootstrap-server %KAFKA_BROKER% --describe --topic %COFFEE_ORDERS_TOPIC% ^| findstr "PartitionCount"') do (
        set CURRENT_PARTITIONS=%%i
    )
    
    if !CURRENT_PARTITIONS! LSS %PARTITIONS% (
        echo Зміна кількості партицій для теми %COFFEE_ORDERS_TOPIC% на %PARTITIONS%...
        kafka-topics.bat --bootstrap-server %KAFKA_BROKER% --alter --topic %COFFEE_ORDERS_TOPIC% --partitions %PARTITIONS%
    ) else (
        echo Тема %COFFEE_ORDERS_TOPIC% вже має !CURRENT_PARTITIONS! партицій, що більше або дорівнює бажаній кількості %PARTITIONS%.
    )
) else (
    echo Створення теми %COFFEE_ORDERS_TOPIC% з %PARTITIONS% партиціями...
    kafka-topics.bat --bootstrap-server %KAFKA_BROKER% --create --topic %COFFEE_ORDERS_TOPIC% --partitions %PARTITIONS% --replication-factor %REPLICATION_FACTOR%
)

REM Перевірка наявності теми processed_orders
echo Перевірка наявності теми %PROCESSED_ORDERS_TOPIC%...
kafka-topics.bat --bootstrap-server %KAFKA_BROKER% --list | findstr /B /C:"%PROCESSED_ORDERS_TOPIC%" > nul
if %ERRORLEVEL% == 0 (
    echo Тема %PROCESSED_ORDERS_TOPIC% вже існує.
    
    REM Отримання поточної кількості партицій
    for /f "tokens=2" %%i in ('kafka-topics.bat --bootstrap-server %KAFKA_BROKER% --describe --topic %PROCESSED_ORDERS_TOPIC% ^| findstr "PartitionCount"') do (
        set CURRENT_PARTITIONS=%%i
    )
    
    if !CURRENT_PARTITIONS! LSS %PARTITIONS% (
        echo Зміна кількості партицій для теми %PROCESSED_ORDERS_TOPIC% на %PARTITIONS%...
        kafka-topics.bat --bootstrap-server %KAFKA_BROKER% --alter --topic %PROCESSED_ORDERS_TOPIC% --partitions %PARTITIONS%
    ) else (
        echo Тема %PROCESSED_ORDERS_TOPIC% вже має !CURRENT_PARTITIONS! партицій, що більше або дорівнює бажаній кількості %PARTITIONS%.
    )
) else (
    echo Створення теми %PROCESSED_ORDERS_TOPIC% з %PARTITIONS% партиціями...
    kafka-topics.bat --bootstrap-server %KAFKA_BROKER% --create --topic %PROCESSED_ORDERS_TOPIC% --partitions %PARTITIONS% --replication-factor %REPLICATION_FACTOR%
)

echo Налаштування тем Kafka завершено.
endlocal
