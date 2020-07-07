package templates

// ResourceNotFound : HTML template for default 404 response
var ResourceNotFound = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    html, body {
      width: 100%;
      height: 100%;
    }

    body {
      margin: 0;
      font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', sans-serif;
      text-rendering: optimizeLegibility;
      -webkit-font-smoothing: antialiased;
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: center;
    }

    h1 {
      font-size: 1.25rem;
      font-weight: 500;
      margin: 0 0 0.5rem;
    }
    
    p {
      margin: 0;
      color: gray;
    }
  </style>
</head>
<body>
  <h1>404</h1>
  <p>Requested resource could not be found</p>
</body>
</html>
`
