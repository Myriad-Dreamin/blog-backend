<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <link rel="stylesheet" href="{{ URL::asset('css/navigator.css') }}">
    <link rel="stylesheet" href="{{ URL::asset('css/main.css') }}">
    <title>Article</title>
</head>
<body>
    <div class="navigator">
		<ul>
			<li><a href="">Home</a></li>
			<li><a href="">Article</a></li>
			<li><a href="">Code</a></li>
			<li><a href="">Mass Article</a></li>
			<li><a href="">Mass Music</a></li>
			<li><a href="">Mass Picture</a></li>
			<li style="float: right;"><a href="">Login</a></li>
		</ul>
        <div class="clear"></div>
    </div>
    
    <div>
    @yield('main_content')
    </div>
</body>
</html>