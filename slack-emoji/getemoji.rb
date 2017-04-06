require "fileutils"
require "open-uri"

require "dotenv"
require "slack"

Dotenv.load

Slack.configure do |config|
  config.token = ENV["SLACK_API_TOKEN"]
end

dir = File.expand_path("../emojis", __FILE__)
FileUtils.mkdir_p(dir) unless FileTest.directory?(dir)

client = Slack::Web::Client.new
client.emoji_list.emoji.each do |name, uri|
  next if uri =~ /^alias:/
  ext = File.extname(uri)
  path = File.join(dir, name + ext)
  next if File.exist?(path)

  puts "Download #{name}..."
  open(uri) do |emoji|
    open(path, "wb") do |f|
      f.write(emoji.read)
    end
  end
end
