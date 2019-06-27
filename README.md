Биржа состоит из 3-х компонентов:
- Клиент
- Брокер
- Биржа

EXCHANGE-SERVER < --- grpc --- > BROKER <--- rest --- > CLIENT
                                          (not done)

## DONE
exchange-server
exchange-server GRPC testing with bufconn

## TODO
broker
client

load test sniper
pprof
-race
exchange-server write results to db

## Биржа

В проекте использоваются исторические данные реальных торгов по фьючерсным контрактам на бирже РТС за 2018-05-18 - https://cloud.mail.ru/public/3faF/W8q4nn7bB

Если на биржу ставится заявка на покупку или продажу, то она ставится в очередь и когда цена доходит до неё и хватает объёма - заявка исполняется, брокеру уходит соответствующее уведомление. Если не хватает объёма, то заявка исполняется частично, брокеру так же уходит уведомление.
Если несколько участников поставилос заявку на одинаковый уровеньт цены, то исполняются в порядке добавления.

Помимо исполнения сделок биржа транслирует цену инструментов всем подключенным брокерам. Список транслируемых инструментов берётся из конфига, сами цены - из файла.

Формат обмена данными с брокером - protobuf через GRPC

## Брокер (NOT DONE)

Брокер предсотавляет своим клиентам доступ на биржу.
У неё есть список клиентов, которые могут взаимодействовать посредством неё с биржей, так же она хранит количество их позиицй и историю сделок.

Брокер аггрегирует внутри себя информацию от биржи по ценовым данным, позволяя клиенту посмотреть историю. По-умолчанию, хранится история за последнеи 5 минут (300 секунд).

Брокер предоставляет клиентам JSON-апи (REST или JSON-RPC), через который им доступныы следующие возможности:
* посмотреть свои позиции и баланс - возвращает баланс + список заявок ( слайс структур ), может быть преобразовано в таблицу на хтмл
* отправить на биржу заявку на покупку или продажу тикера ( то что вы видите на скрине у клиента )
* отменить ранее отправленную заявку - принимает ИД заявки
* посмотреть последнюю истории торгов - возвращает слайс структур, может быть преобразовано в таблицу на хтмл


## Клиент (NOT DONE)

Клиент - биржевой терминал на хтмл, который нам позволяет покупать-продавать и смотреть свою историю, пользователь АПИ брокера



