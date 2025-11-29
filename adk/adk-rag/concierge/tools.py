import datetime


def now_tool():
    """現在の時刻を返します。"""
    return datetime.datetime.now().strftime("%Y年%m月%d日 %H時%M分%S秒")


def get_weather(city: str) -> dict:
    """
    指定された都市の現在の天気情報を返します。

    Args:
        city (str): 天気情報を取得したい都市名。英名で指定してください。 tokyo, new york など

    Returns:
        dict: 天気情報を含む辞書。成功した場合は天気レポート、失敗した場合はエラーメッセージが含まれます。
    """
    if city.lower() == "new york":
        return {
            "status": "success",
            "report": (
                "The weather in New York is sunny with a temperature of 25 degrees"
                " Celsius (77 degrees Fahrenheit)."
            ),
        }
    elif city.lower() == "tokyo":
        return {
            "status": "success",
            "report": (
                "The weather in Tokyo is cloudy with a temperature of 22 degrees"
                " Celsius (72 degrees Fahrenheit)."
            ),
        }
    else:
        return {
            "status": "error",
            "error_message": f"Weather information for '{city}' is not available.",
        }
