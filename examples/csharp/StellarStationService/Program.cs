var satId = "297";

var item = Stellarstation.Configuration.FromFile("./config.json");

var chan = Stellarstation.Auth.GenerateChannel(item.apiAddress, item.apiKeyPath);
var client = new Stellarstation.Client(chan);

if (item.satellites != null) {
    foreach (var satellite in item.satellites) {
        var res = client.ListUpcomingAvailablePasses(satId, DateTime.UtcNow, DateTime.UtcNow.Add(TimeSpan.FromDays(3)));
        Console.Write(res.ToString());
    }
}


