#!/bin/bash

# Скрипт для сравнения вывода mygrep и оригинального grep (версия для Bash)
# Используется для демонстрации корректности распределенной обработки.

TEST_FILE="testdata.txt"
PORT1=6001
PORT2=6002
COORDINATOR_PORT=6000

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 1. Собираем исполняемые файлы
echo -e "${CYAN}Building binaries...${NC}"
go build -o coordinator coordinator.go
go build -o worker worker.go

# 2. Запускаем воркеров в фоновом режиме
echo -e "${CYAN}Starting workers...${NC}"
./worker -port :$PORT1 -callback "http://localhost:$COORDINATOR_PORT" > worker1.log 2>&1 &
W1_PID=$!
./worker -port :$PORT2 -callback "http://localhost:$COORDINATOR_PORT" > worker2.log 2>&1 &
W2_PID=$!

sleep 2

# Функция очистки при выходе
cleanup() {
    echo -e "\n${CYAN}Stopping workers (PIDs: $W1_PID, $W2_PID)...${NC}"
    kill $W1_PID $W2_PID 2>/dev/null
    rm -f coordinator worker worker1.log worker2.log
}

trap cleanup EXIT

compare_grep() {
    local title=$1
    local mygrep_extra_args=$2
    local pattern=$3
    local grep_flags=$4

    echo -e "\n${YELLOW}--- Testing: $title ---${NC}"

    # Запуск нашего mygrep
    # Фильтруем логи координатора, оставляя только результаты поиска
    my_result=$(./coordinator --pattern "$pattern" \
        --servers "http://localhost:$PORT1,http://localhost:$PORT2" \
        --file "$TEST_FILE" \
        --quorum 2 \
        --port ":$COORDINATOR_PORT" \
        $mygrep_extra_args 2>&1 | \
        grep -v "^\[COORDINATOR\]" | \
        grep -v "^[0-9]\{4\}/[0-9]\{2\}/[0-9]\{2\}" | \
        grep -v "Accepted|Posted|Received response" | \
        grep -v "^Usage of" | \
        grep -v "^  -" | \
        sed '/^[[:space:]]*$/d' | \
        sort)

    # Запуск оригинального grep
    expected_result=$(grep $grep_flags "$pattern" "$TEST_FILE" | sort)

    if [ "$my_result" == "$expected_result" ]; then
        echo -e "${GREEN}SUCCESS: Results match!${NC}"
    else
        echo -e "${RED}FAILURE: Results differ!${NC}"
        echo -e "${YELLOW}Expected:${NC}"
        echo "$expected_result"
        echo -e "${YELLOW}Got:${NC}"
        echo "$my_result"
        
        # Показываем diff если есть различия
        diff <(echo "$my_result") <(echo "$expected_result")
    fi
}

# Тест 1: Простой поиск (case-sensitive)
compare_grep "Simple Search" "--chunksize 5" "ERROR" ""

# Тест 2: Поиск без учета регистра
compare_grep "Ignore Case" "--chunksize 5 --i" "error" "-i"

# Тест 3: Инвертированный поиск
compare_grep "Invert Match" "--chunksize 5 --v" "normal" "-v"

# Тест 4: Поиск с контекстом (After)
compare_grep "Context After" "--chunksize 10 --A 1" "failure" "-A 1"

echo -e "\n${GREEN}All tests finished.${NC}"
