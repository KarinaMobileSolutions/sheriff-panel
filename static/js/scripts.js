(function() {
    var app = angular.module('scripts', []);

    plots = [];
    plotsData = {};

    (function() {
        var conn = new WebSocket("ws://127.0.0.1:8080/ws");
        conn.onclose = function(evt) {
            $('#errormodal').modal('show');
        };
        conn.onmessage = function(evt) {
            var result = $.parseJSON(evt.data)[1].split(':');
            plotsData[result[0]] = plotsData[result[0]].slice(1);
            plotsData[result[0]].push([result[1] * 1000, result[2]]);
            plots[result[0]].setData([plotsData[result[0]]]);
            plots[result[0]].setupGrid();
            plots[result[0]].draw();
        };
    })();


    timezoneJS.timezone.zoneFileBasePath = "static/tz";
    timezoneJS.timezone.defaultZoneFile = [];
    timezoneJS.timezone.init({async: false});

    app.directive('scriptInfo', function() {
        return {
            restrict: 'E',
            scope: {
                name: '@name'
            },
            templateUrl: 'static/tmpl/script-info.html',
            controller: ['$http', '$scope', function($http, $scope) {
                $http.get('/scripts/'+$scope.name).success(function(data) {
                    $scope.details = data;
                });
            }],
            controllerAs: 'script',
        };
    });

    app.directive('scriptChart', function() {
        return {
            restrict: 'E',
            scope: {
                name: '@name'
            },
            templateUrl: 'static/tmpl/script-chart.html',
            controller: ['$http', '$scope', function($http, $scope) {
                $http.get('/scripts/chart/'+$scope.name).success(function(data) {
                    var chartData = [];

                    $.each(data, function(index, key) {
                        var value = index.split(':')[1];

                        chartData.push([key * 1000, value]);
                    });

                    plots[$scope.name] = $.plot("#chart"+$scope.name, [chartData], {series:{shadowSize:0}, xaxis: {mode: "time", timezone:"browser"}});
                });
            }],
            controllerAs: 'script',
        };
    });
})();
