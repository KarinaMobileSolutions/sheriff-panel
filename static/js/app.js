(function() {
    var app = angular.module('Sheriff', ['scripts']);

    app.controller('ChartController', ['$http', function($http) {
        this.charts = [];
        object = this;

        $http.get('/scripts').success(function(data) {
            data.forEach(function (value, key) {
                object.charts.push({'script': {'name': value}});
            });
        });
    }]);
})();
