using Newtonsoft.Json;

namespace Stellarstation {
    public static class Configuration {
        public static Config FromFile(string filePath) {
            using (StreamReader r = new StreamReader(filePath)) {
                string json = r.ReadToEnd();
                var item = JsonConvert.DeserializeObject<Config>(json);
                if (item == null || item.apiAddress == null || item.apiKeyPath == null || item.groundStations == null || item.satellites == null) {
                    throw new ArgumentException("Error unmarshalling the config file");
                }

                return item;
            }
        }
    }

    public class Config {
        [JsonProperty("api_address")]
        public string? apiAddress;
        [JsonProperty("api_key_path")]
        public string? apiKeyPath;
        [JsonProperty("ground_stations")]
        public List<GroundStation>? groundStations;
        [JsonProperty("satellites")]
        public List<Satellite>? satellites;
    }

    public class GroundStation {
        [JsonProperty("id")]
        public int Id;
        [JsonProperty("name")]
        public string? Name ;
    }

    public class Satellite {
        [JsonProperty("id")]
        public int Id;
        [JsonProperty("name")]
        public string? Name;
    }
}