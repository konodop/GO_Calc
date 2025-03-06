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
<h1>инструкция для запуска проекта:<h1\>
<h2>шаг 1 <h2\><h4>скачать веб сервис через git clone
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>git clone https://github.com/konodop/GO_project_second.git</pre></div>
или просто скачать и распаковать .zip файл из гитхаба если не установлен git
<h4\>
<h2>шаг 2<h2\><h4> Запуск сервера. В основной папке с помщью терминала либо git bash ввести команду:<h4\>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>go run ./cmd/main.go</pre></div>
<h2>шаг 3<h2\><h4> Отправка POST-запроса через curl. Снова открываем командную строку с помощью которой можно будет отправлять запросы например:<h4\>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>curl -X POST http://localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d "{\"expression\": \"1+1\"}"</pre></div>
    Ответ:
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>{"id":"1"}</pre></div>
<h3>Можно подставлять другие значения и проверять их<h3\>
<hr><hr\>
<h1>Эндпоинты<h1\>
<h3>Основной для отправки задач<h3\>
<div class="highlight highlight-source-shell" dir="auto"><pre>/api/v1/calculate</pre></div>
<h3>Для проверки отправленных задач<h3\>
<div class="highlight highlight-source-shell" dir="auto"><pre>/api/v1/expressions</pre></div>
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
<h1>Фронтенд?</h1>
<h3>Вы можете перейти по вышеуказанным эндпоинтам в браузере (типо так http://127.0.0.1:8080/api/v1/expressions) и получите ответы от сервера в виде Json, но лучше не переходить на эндпоинт с тасками чтобы случайно не забрать у агента задачу🫥
</h3>
<hr><hr\>
<h1>Тесты</h1>
<h3>Нужно зайти в попку с кодом который нужно проверить и написать в терминал go test -v</h3>
<hr><hr\>
<h1>Состав проекта<h1\>
<h3>
<ul>
    <li>cmd/main.go______________________________файл для запуска приложения и сервера</li>
    <li>internal\aplication\application.go_______файл сервера и обработки ошибок</li>
    <li>pkg/calculation/calculation.go___________файл самого калькулятора</li>
    <li>pkg/calculation/calculation.go___________файл тестов калькулятора</li>
    <li>go.mod___________________________________модуль соединяющий остальные файлы</li>
    </ul>
<h3\>
