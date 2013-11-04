using System;
using ServiceStack.Text;
using ServiceStack.Text.Json;
using System.Collections.Generic;
using Couchbase;
using Couchbase.Extensions;
using Enyim.Caching.Memcached;
using System.Threading.Tasks;

namespace CouchbaseWoWTests
{
	public class MainClass
	{
		private static int _startId = 40000;
		private static int _endId = 80010;
		private static string _baseUrl = "http://eu.battle.net/api/wow/item/{0}";

		private static CouchbaseClient _client;

		public class WoWItem
		{
			public string Id { get; set; }
			public string Name {get;set;}
			public string Status {get;set;}
		}

		public static void Main (string[] args)
		{
			Console.WriteLine ("Creating client...");
			_client = new CouchbaseClient ();

			Console.WriteLine ("Importing Documents {0} through {1}...", _startId, _endId);
			Import ();
		}

		public static void Import()
		{
			for (int i = _startId; i <= _endId; i++)
			{
				var parts = DownloadWoWItem (i);
				if (parts == null)
					continue;

				if (i % 100 == 0)
					Console.WriteLine ("{0}", i);

				var key = string.Format ("item_{0}", parts.Item1.Id);
				_client.Store (StoreMode.Set, key, parts.Item2);

			}
		}

		public static Tuple<WoWItem, string> DownloadWoWItem(int id)
		{
			var url = string.Format (_baseUrl, id);
			var wc = new System.Net.Http.HttpClient ();
			try
			{
				var result = wc.GetStringAsync (url).Result;
				var wowItem = JsonSerializer.DeserializeFromString<WoWItem> (result);

				if (wowItem.Status == "nok")
					return null;

				var ret = new Tuple<WoWItem,string>(wowItem, result);
				return ret;
			}
			catch
			{
				return null;
			}

		}
	}
}
