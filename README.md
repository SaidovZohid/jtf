# [JTF (Just Transfer File)](https://jtf.zohiddev.me)
<img src="https://jtf.zohiddev.me/assets/logo.png" align="right" width="245">

## Overview
JTF (Just Transfer Files) is a command-line application written in Golang, designed to simplify file transfers using SSH and enable downloading files via HTTP. With JTF, developers can seamlessly send files from their terminal using SSH and generate links for download page, direct download link, and delete file link.

## Features
* Effortless file transfers via SSH from the terminal.
* Generate links for download page, direct download link, and delete file link.
* Various options for customizing file transfers and generating links.
* Fast and secure file transfer using SSH.

## Usage
1. Sending a file via SSH and generating links:
```bash
# Sending a file via SSH and generating links
ssh jtf.zohiddev.me -p 2222 < dump.json
        🌟✨ Welcome to JTF! ✨🌟

# Output:
🌟🔒 Detected verified user domain https://zohid.jtf.zohiddev.me 🔒🌟

Download link:
	https://zohid.jtf.zohiddev.me/2g3pev8

Direct download link:
	https://direct.jtf.zohiddev.me/2g3pev8

Delete file link:
	https://delete.jtf.zohiddev.me/2g3pev8

⏳ Please hurry! Your link will expire in 15 minutes. After that, the session will automatically close, and the link will become invalid. Let's patiently wait for the download to commence... 🕒

🛎  Exciting news! 📥 Your file downloaded by someone. 🎉✨
```

2. Sending a file via SSH with additional options:
```bash
ssh jtf.zohiddev.me -p 2222 from="Alex" < dump.json # Use the "from=" option to set a custom name for the download page. 
ssh jtf.zohiddev.me -p 2222 filename="just.json" < dump.json # Customize the "filename=" parameter to give your downloaded file a unique name.
ssh jtf.zohiddev.me -p 2222 msg="This file is for you" < dump.json # Add a personalized "msg=" to include a special message along with the file.
ssh jtf.zohiddev.me -p 2222 t=2 < file.txt # You can change the download availability time by specifying the "t" option (0 < sv < 60) during file upload.
ssh jtf.zohiddev.me -p 2222 filename="just.json" msg="This file is for you" from="Alex" t=10 < dump.json # All in one command 
```

## Generated Links
Upon successful file transfer, JTF will generate the following links:
1. **Download Page Link**: https://zohid.jtf.zohiddev.me/2g3pev8
2. **Direct Download Link**: https://zohid.jtf.zohiddev.me/2g3pev8
3. **Delete File Link**: https://zohid.jtf.zohiddev.me/2g3pev8

## Contributing
We welcome contributions from the community! If you have any ideas to improve JTF or encounter any issues, please refer to our [CONTRIBUTING.md](https://github.com/SaidovZohid/jtf/blob/main/CONTRIBUTING.md) file for detailed guidelines on how to contribute, including information on code standards, testing, and pull request submission.

## License
JTF is an open-source software licensed under the [MIT License](https://github.com/SaidovZohid/jtf/blob/main/LICENSE).

### Securely share files without leaving the terminal command-line environment. Happy coding!