<h1 align="center">веб-сервис<h1\>
    <h2>Описание:</h2>
    <h3>пользователь отправляет арифметическое выражение по HTTP в формате JSON, а сервис отправляет его на выполнение.
    строка-выражение состоит из односимвольных идентификаторов и знаков арифметических действий.
    Входящие данные - цифры(рациональные), операции +, -, *, /, операции приоритезации ( ). В случае ошибки записи выражения приложение выдает ошибку.</h3>
    <h3>В сервисе есть две части:
        <ul dir="auto">
        <ol>1. Оркестратор, отправляющий агенту задачи</ol>
        <ol>2. Агент (демон) выполняет простые задачи от оркестратора, а также может выполнять их параллельно.</ol>
        </ul>
    У сервиса основной endpoint с url-ом /api/v1/calculate
    </h3>
    <h1>Требования:<h1\>
    <h2>
    <ul>
    <li>golang, можно скачать на https://go.dev/dl/</li>
    <li>(необязательно) git, можно скачать на https://git-scm.com/downloads</li>
    </ul>
    </h2>
<hr><hr\>
<h2>Схема работы:</h2>
        <img src="https://github.com/konodop/Pics/blob/go/image.png">,
<hr><hr\>
<h1>инструкция для запуска проекта:<h1\>
<h2>шаг 1 <h2\><h4>скачать веб сервис через git clone
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>git clone https://github.com/konodop/GO_project_second.git</pre></div>
или просто скачать и распаковать .zip файл из гитхаба если не установлен git
<h4\>
<h2>шаг 2<h2\><h4> Запуск сервера. В основной папке с помщью терминала либо git bash ввести команду:<h4\>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>go run ./cmd/orchestrator/main.go</pre></div>
<h2>шаг 3<h2\><h4> Запуск агента. В той же папке с помщью терминала либо git bash ввести команду (не закрывать прошлый шаг, а в новом терминале запускать):<h4\>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>go run ./cmd/agent/main.go</pre></div>
<h2>шаг 4<h2\><h4> Отправка POST-запроса через curl. Снова открываем командную строку с помощью которой можно будет отправлять запросы например:<h4\>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>curl -X POST http://localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d "{\"expression\": \"1+1\"}"</pre></div>
    Ответ:
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>{"id":"1"}</pre></div>
<h3>Можно подставлять другие значения и проверять их<h3\>
<h2>шаг 5<h2\><h4>Смотрим ответ этой задачи<h4\>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>curl -X GET http://localhost:8080/api/v1/expressions?id=1</pre></div>
    Ответ:
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>{"id":1,"status":"ended","result":"2.000000"}</pre></div>
<hr><hr\>
<h1>Эндпоинты<h1\>
<h3>Основной для отправки задач<h3\>
<div class="highlight highlight-source-shell" dir="auto"><pre>/api/v1/calculate</pre></div>
<h3>Для проверки отправленных задач<h3\>
<div class="highlight highlight-source-shell" dir="auto"><pre>/api/v1/expressions</pre></div>
<h3>Также он умеет выводить по айди отправленную задачу<h3\>
<div class="highlight highlight-source-shell" dir="auto"><pre>/api/v1/expressions?id={сюда писать id}</pre></div>
<h3>Получение, а также отправка задачи для выполнения<h3\>
<div class="highlight highlight-source-shell" dir="auto"><pre>/api/internal/task</pre></div>
<hr><hr\>
<h1>Примеры запросов:<h1\>
<h2>1</h2>
<h4>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>curl -X POST http://localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d "{\"expression\": \"1-(1000*10)\"}"</pre></div>
Ответ:
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>{"id":"1"}</pre><div class="zeroclipboard-container"></div>
Статус 201 (Задча отправлена), всë правильно.
</h4>
<h2>2</h2>
<h4>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>curl -X POST http://localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d "{\"expression\": \"1/0\"}"</pre></div>
Ответ:
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>{"id":"2"}</pre><div class="zeroclipboard-container"></div>
Статус 201 (Задча отправлена), но оркестратор потом выявит ошибку.
</h4>
<h2>Проверка отправленных задач</h2>
<h4>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>curl -X GET http://localhost:8080/api/v1/expressions</pre></div>
Ответ:
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>{"expressions":[{"id":1,"status":"ended","result":"-9999.000000"},{"id":2,"status":"Cannot divide by 0","result":"NULL"}]}</pre><div class="zeroclipboard-container"></div>
Видно, что первая задча выполнилась, а во второй есть ошибка
<hr><hr\>
<h1>Фронтенд</h1>
<h4>позволяет по url (http://localhost:8081/) видеть отправленные задачи. Чтобы его увидеть запустите front.exe или бинарный файл или напишите эту команду в папке cmd/front:<h4\>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>go run ./front.go</pre></div>
</h3>
<hr><hr\>
<h1>Тесты</h1>
<h3>Нужно зайти в попку с кодом который нужно проверить и написать в терминал go test -v</h3>
<hr><hr\>
<h1>Состав проекта<h1\>
<h3>
<ul>
    <li>cmd/orchestrator/main.go________________файл для запуска оркестратора</li>
    <li>cmd/agent/main.go_______________________файл для запуска агента</li>
    <li>cmd/agent/front.go_______________________файл для запуска веб дизайна</li>
    <li>cmd/agent/html.html_______________________файл с форматом html</li>
    <li>internal/aplication/orchestrator.go_______файл оркестратора</li>
    <li>internal/aplication/orchestrator_test.go__файл тестов оркестратора</li>
    <li>pkg/calculation/agent.go_________________файл агента</li>
    <li>pkg/calculation/agent_test.go____________файл тестов агента</li>
    <li>go.mod____________________________________модуль соединяющий остальные файлы</li>
    <li>cmd/agent/main.go_______________________файл для запуска агента</li>
    <li>cmd/bin________________________________папка с бинарниками для запуска всех частей</li>
    <li>cmd/exe________________________________папка с exe файлами для запуска всех частей на виндовс</li>
    </ul>
<h3\>
