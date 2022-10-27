var apiAddress = "https://api.stellarstation.com";
var apiKey = "./api-key.json";
var satId = "297";

var chan = Stellarstation.Auth.GenerateChannel(apiAddress, apiKey);

var client = new Stellarstation.Client(chan);

var res = client.ListUpcomingAvailablePasses(satId, DateTime.UtcNow, DateTime.UtcNow.Add(TimeSpan.FromDays(3)));
Console.Write(res.ToString());
