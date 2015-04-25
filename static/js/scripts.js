(function() {
    var app = angular.module('scripts', []);

    plots = [];
    plotsData = {};

    app.run(['$location', function($location) {
        var conn = new WebSocket("ws://"+$location.host()+":"+$location.port()+"/ws");
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
    }]);


    timezoneJS.timezone.zoneFileBasePath = "static/tz";
    timezoneJS.timezone.defaultZoneFile = [];
    timezoneJS.timezone.init({async: false});

    app.directive('scriptInfoModal', function() {
        return {
            restrict: 'E',
            scope: {
                name: '@name'
            },
            templateUrl: 'static/tmpl/script-info-modal.html?v=0.1',
            controller: ['$http', '$scope', function($http, $scope) {
                $http.get('/scripts/'+$scope.name).success(function(data) {
                    $scope.details = data;
                });
            }],
            controllerAs: 'script',
        };
    });


    app.directive('scriptInfo', function() {
        return {
            restrict: 'E',
            scope: {
                name: '@name'
            },
            templateUrl: 'static/tmpl/script-info.html?v=0.1',
        };
    });

    app.directive('scriptChart', function() {
        return {
            restrict: 'E',
            scope: {
                name: '@name'
            },
            templateUrl: 'static/tmpl/script-chart.html?v=0.1',
            controller: ['$http', '$scope', function($http, $scope) {
                this.period = 'hour';

                var chart = this;

                this.draw = function() {
                    $http.get('/scripts/chart/'+$scope.name+'?period='+chart.period).success(function(data) {
                        var chartData = [];

                        $.each(data, function(index, key) {
                            var value = index.split(':')[1];

                            chartData.push([key * 1000, value]);
                        });

                        plotsData[$scope.name] = chartData;

                        plots[$scope.name] = $.plot("#chart"+$scope.name, [chartData], {series:{shadowSize:0}, xaxis: {mode: "time", timezone:"browser"}});
                    });
                };

                this.draw();
            }],
            controllerAs: 'script',
        };
    });
})();
