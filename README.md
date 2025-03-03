<h1>веб-сервис<h1\>
    <h2>Описание:</h2>
    <h3>пользователь отправляет арифметическое выражение по HTTP в формате JSON и получает в ответ его результат в виде JSON.
    строка-выражение состоит из односимвольных идентификаторов и знаков арифметических действий.
    Входящие данные - цифры(рациональные), операции +, -, *, /, операции приоритезации ( ). В случае ошибки записи выражения приложение выдает ошибку.
    У сервиса endpoint с url-ом /api/v1/calculate
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
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>git clone https://github.com/konodop/Project_go.git</pre></div>
или просто скачать и распаковать .zip файл из гитхаба если не установлен git
<h4\>
<h2>шаг 2<h2\><h4> Запуск сервера. В основной папке с помщью терминала либо git bash ввести команду:<h4\>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>go run ./cmd/main.go</pre></div>
<h2>шаг 3<h2\><h4> Отправка POST-запроса через curl. Снова открываем командную строку с помощью которой можно будет отправлять запросы например:<h4\>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>curl -X POST http://localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d "{\"expression\": \"1+1\"}"</pre></div>
    Ответ:
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>{"result":"2.000000"}</pre></div>
<h3>Можно подставлять другие значения в expression и проверять их<h3\>
<hr><hr\>
<h1>Примеры запросов:<h1\>
<h2>1</h2>
<h4>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>curl -X POST http://localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d "{\"expression\": \"1-(1000*10)\"}"</pre></div>
Ответ:
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>{"result":"-9999.000000"}</pre><div class="zeroclipboard-container"></div>
Статус 200 (ОК), всë правильно.
</h4>
<h2>2</h2>
<h4>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>curl -X POST http://localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d "{\"expression\": \"1/0\"}"</pre></div>
Ответ:
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>{"error":"Expression is not valid"}</pre><div class="zeroclipboard-container"></div>
Ошибка 422 (Unprocessable Entity), так как нельзя делить на ноль.
</h4>
<h2>Также есть ошибка 500 (Internal Server Error)</h2>
<h4>
<div class="highlight highlight-source-shell notranslate position-relative overflow-auto" dir="auto"><pre>{"error": "Internal Server Error"}</pre><div class="zeroclipboard-container"></div>
Появляется она только если произойдёт чудо и в сервере true будет равнятся false
</h4>
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
