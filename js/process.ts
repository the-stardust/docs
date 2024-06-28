// @ts-nocheck


interface DetectResult {
    status: 'success' | 'error'
    message: string | undefined
    data: {
        phone: string,
        image_list: {
            id: string,
            image: string
        }
    }
}

class Detect {
    private readonly dictionary: any
    private grayList: any[] = []
    private adaptiveList: any[] = []
    private originList: any[] = []
    private idsIndex = 0
    private ids = -1
    private idsCount = 0
    private idsPoint = {}
    private idsMinX = 0
    private idsMinY = 0
    private idsMaxX = 0
    private idsMaxY = 0
    private ZOOM_SCALAR = 1
    private SL_MIN_WIDTH = 1300
    private SL_MAX_WIDTH = 1350
    private SL_MIN_HEIGHT = 10
    private SL_MAX_HEIGHT = 40
    public static tempGray = null
    public static xingce_value = 0.4


    constructor() {
        this.dictionary = new cv.aruco_Dictionary(cv.DICT_4X4_50);
    }

    static resize(input, width, height) {
        let output = new cv.Mat()
        // console.log(input, width, height)
        cv.resize(input, output, new cv.Size(width, height))
        return output
    }


    private static fourPointsTransform(mat, points) {
        let br = points.reduce((prev, current) => (prev.x + prev.y > current.x + current.y) ? prev : current)
        let tl = points.reduce((prev, current) => (prev.x + prev.y < current.x + current.y) ? prev : current)
        let other = points.filter(obj => (obj.x !== br.x || obj.y !== br.y) && (obj.x !== tl.x || obj.y !== tl.y))
        let bl = other.reduce((prev, current) => (prev.y > current.y) ? prev : current)
        let tr = other.reduce((prev, current) => (prev.x > current.x) ? prev : current)
        let widthA = Math.sqrt(Math.pow((br.x - bl.x), 2.0) + Math.pow((br.y - bl.y), 2.0))
        let widthB = Math.sqrt(Math.pow((tr.x - tl.x), 2.0) + Math.pow((tr.y - tl.y), 2.0))
        let heightA = Math.sqrt(Math.pow((tr.x - br.x), 2.0) + Math.pow((tr.y - br.y), 2.0))
        let heightB = Math.sqrt(Math.pow((tl.x - bl.x), 2.0) + Math.pow((tl.y - bl.y), 2.0))
        let maxHeight = heightA > heightB ? heightA : heightB
        let maxWidth = widthA > widthB ? widthA : widthB
        let srcTri = cv.matFromArray(4, 1, cv.CV_32FC2, [tl.x, tl.y, tr.x, tr.y, br.x, br.y, bl.x, bl.y]);
        let dstTri = cv.matFromArray(4, 1, cv.CV_32FC2, [0, 0, maxWidth - 1, 0, maxWidth - 1, maxHeight - 1, 0, maxHeight]);
        let M = cv.getPerspectiveTransform(srcTri, dstTri);
        let four = new cv.Mat()

        let dsize = new cv.Size(maxWidth, maxHeight);
        cv.warpPerspective(mat, four, M, dsize, cv.INTER_LINEAR, cv.BORDER_CONSTANT);
        M.delete()

        return four
    }

    private static euclideanDistance(p1, p2) {
        const dx = p1.x - p2.x;
        const dy = p1.y - p2.y;
        return Math.sqrt(dx * dx + dy * dy);
    }


    private detectMarkersInRects(mat: cv.Mat, rects: cv.Rect[]) {
        const markerIds = new cv.Mat();
        const markerCorners = new cv.MatVector();

        try {
            for (const rect of rects) {
                const roi = mat.roi(rect);
                // cv.adaptiveThreshold(roi, roi, 255, cv.ADAPTIVE_THRESH_MEAN_C, cv.THRESH_BINARY, 5, 1)

                // let canv = document.createElement('canvas');
                // cv.imshow(canv, roi)
                // document.body.appendChild(canv)
                cv.detectMarkers(roi, this.dictionary, markerCorners, markerIds);
                roi.delete();

                if (markerCorners.size() !== 0) {
                    const corners = markerCorners.get(0);
                    const xData = corners.data32F.filter((_, i) => i % 2 === 0);
                    const yData = corners.data32F.filter((_, i) => i % 2 !== 0);

                    const minX = Math.min(...xData) + rect.x;
                    const maxX = Math.max(...xData) + rect.x;
                    const minY = Math.min(...yData) + rect.y;
                    const maxY = Math.max(...yData) + rect.y;

                    // 返回真实坐标
                    return {
                        id: markerIds.data32S[0],
                        x: minX,
                        y: minY,
                        width: maxX - minX,
                        height: maxY - minY,
                    };
                }
            }
        } finally {
            markerIds.delete();
            markerCorners.delete();
        }

        return null;
    }

    private preprocess(imageList: any[]) {
        let i = 0
        for (let img of imageList) {
            let gray = new cv.Mat()
            let adaptive = new cv.Mat()
            cv.cvtColor(img, gray, cv.COLOR_RGBA2GRAY, 0)
            cv.adaptiveThreshold(gray, adaptive, 255, cv.ADAPTIVE_THRESH_MEAN_C, cv.THRESH_BINARY, 155, 1)
            this.grayList.push(gray)
            this.adaptiveList.push(adaptive)
            // let markerIds = new cv.Mat();
            // let markerCorners = new cv.MatVector();
            let width = img.cols
            let height = img.rows

            let result = this.detectMarkersInRects(gray, [
                new cv.Rect(width - 130, 530, 130, 130),
                new cv.Rect(0, height - 660, 130, 130),

                new cv.Rect(500, 0, 200, 200),
                new cv.Rect(width - 200, 500, 200, 200),
                new cv.Rect(0, 500, 200, 200),
                new cv.Rect(width - 500 - 200, height - 200, 200, 200),

            ])
            console.log(result)

            if (result) {
                this.idsIndex = i
                this.ids = result.id
                this.idsCount += 1
                this.idsPoint = {x: result.x, y: result.y}
                this.idsMaxX = result.x + result.width
                this.idsMaxY = result.y + result.height
                this.idsMinX = result.x
                this.idsMinY = result.y
            }

            // let tmpMat = new cv.Mat()
            // cv.resize(gray, tmpMat, new cv.Size(Math.floor(gray.cols / this.ZOOM_SCALAR), Math.floor(gray.rows / this.ZOOM_SCALAR)))
            // cv.detectMarkers(tmpMat, this.dictionary, markerCorners, markerIds);
            // tmpMat.delete()
            //
            // if (markerCorners.size() !== 0) {
            //     let corners = markerCorners.get(0)
            //     this.idsIndex = i
            //     this.ids = markerIds.data32S[0]
            //     console.log(markerIds.data32S)
            //     this.idsCount += 1
            //     this.idsPoint = {x: corners.floatAt(0, 0) * this.ZOOM_SCALAR, y: corners.floatAt(0, 1) * this.ZOOM_SCALAR}
            //     let idsX = [corners.floatAt(0, 0), corners.floatAt(0, 2), corners.floatAt(0, 4), corners.floatAt(0, 6)]
            //     let idsY = [corners.floatAt(0, 1), corners.floatAt(0, 3), corners.floatAt(0, 5), corners.floatAt(0, 7)]
            //
            //     this.idsMaxX = Math.max(...idsX) * this.ZOOM_SCALAR
            //     this.idsMaxY = Math.max(...idsY) * this.ZOOM_SCALAR
            //     this.idsMinX = Math.min(...idsX) * this.ZOOM_SCALAR
            //     this.idsMinY = Math.min(...idsY) * this.ZOOM_SCALAR
            //     corners.delete()
            // }
            // markerIds.delete()
            // markerCorners.delete()
            i += 1
        }
    }

    // 计算矩形范围内，黑色像素占比
    private static calcRectWeight(mat, rect) {
        let count = 0
        for (let x = rect.x; x < rect.x + rect.width; x++) {
            for (let y = rect.y; y < rect.y + rect.height; y++) {
                if (mat.ucharAt(y, x) == 0) {
                    count++
                }
            }
        }

        return count / (rect.width * rect.height)
    }

    // 计算手机号函数，根据不同的phoneOption
    private getPhone(phoneOption) {
        let phone = ''
        let start = {x: phoneOption.rect.x, y: phoneOption.rect.y}
        let step = {x: phoneOption.rect.width / 11, y: phoneOption.rect.height / 10}

        for (let x = 0; x < 11; x++) {
            let score = []
            for (let y = 0; y < 10; y++) {
                let currentX = start.x + step.x * x
                let currentY = start.y + step.y * y

                let rect = new cv.Rect(currentX + 11, currentY + 11, 30, 20)
                let sp1 = new cv.Point(rect.x, rect.y)
                let sp2 = new cv.Point(rect.x + rect.width, rect.y + rect.height)
                let weight = Detect.calcRectWeight(this.adaptiveList[this.idsIndex], rect)
                score.push(weight)
                if (weight > 0.4) {
                    cv.rectangle(this.originList[this.idsIndex], sp1, sp2, [255, 0, 0, 255], 2)
                }
                // console.log()
            }

            console.log(score)
            let maxScore = Math.max(...score)
            if (maxScore < 0.4) {
                phone += 'u'
            } else {
                phone += (score.indexOf(maxScore)).toString()
            }
        }
        // cv.imshow('output', this.originList[this.idsIndex])
        return phone
    }


    private static getPhoneFromCamara(originImage, adaptiveImage, phoneOption) {
        let phone = ''
        let start = {x: phoneOption.rect.x, y: phoneOption.rect.y}
        let step = {x: phoneOption.rect.width / 11, y: phoneOption.rect.height / 10}

        for (let x = 0; x < 11; x++) {
            let score = []
            for (let y = 0; y < 10; y++) {
                let currentX = start.x + step.x * x
                let currentY = start.y + step.y * y
                let rect = new cv.Rect(currentX + 11, currentY + 11, 30, 13)
                let sp1 = new cv.Point(rect.x, rect.y)
                let sp2 = new cv.Point(rect.x + rect.width, rect.y + rect.height)
                let weight = this.calcRectWeight(adaptiveImage, rect)
                cv.rectangle(originImage, sp1, sp2, [255, 0, 0, 255], 2)
                score.push(weight)
            }

            let maxScore = Math.max(...score)
            if (maxScore < 0.4) {
                phone += 'u'
            } else {
                phone += (score.indexOf(maxScore)).toString()
            }
        }

        return phone
    }

    private toHsv(mat) {
        let quesImage = mat
        let bgr = new cv.Mat()
        cv.cvtColor(quesImage, bgr, cv.COLOR_RGBA2BGR)
        let hsv = new cv.Mat()
        cv.cvtColor(bgr, hsv, cv.COLOR_BGR2HSV)
        let hsvOutput = new cv.Mat()


        let low = new cv.Mat(hsv.rows, hsv.cols, hsv.type(), [0, 0, 0, 0]);
        let high = new cv.Mat(hsv.rows, hsv.cols, hsv.type(), [180, 255, 80, 0])
        cv.inRange(hsv, low,  high, hsvOutput)
        let canv = document.createElement('canvas');
        // cv.imshow(canv, hsvOutput)
        // document.body.appendChild(canv)

        mat.delete()
        low.delete()
        high.delete()
        hsv.delete()
        hsvOutput.delete()
        bgr.delete()

        return canv.toDataURL()
    }

    static toHsv(mat) {
        let quesImage = mat
        let bgr = new cv.Mat()
        cv.cvtColor(quesImage, bgr, cv.COLOR_RGBA2BGR)
        let hsv = new cv.Mat()
        cv.cvtColor(bgr, hsv, cv.COLOR_BGR2HSV)
        let hsvOutput = new cv.Mat()


        let low = new cv.Mat(hsv.rows, hsv.cols, hsv.type(), [0, 0, 0, 0]);
        let high = new cv.Mat(hsv.rows, hsv.cols, hsv.type(), [180, 255, 80, 0])
        cv.inRange(hsv, low,  high, hsvOutput)
        let canv = document.createElement('canvas');
        cv.imshow(canv, bgr)
        // document.body.appendChild(canv)

        quesImage.delete()
        low.delete()
        high.delete()
        hsv.delete()
        hsvOutput.delete()
        bgr.delete()

        return canv.toDataURL('image/jpeg', 0.6)
    }

    // 处理申论答题卡
    private detectSlCard() {
        let image_list = []
        let width = this.originList[this.idsIndex].cols
        let height = this.originList[this.idsIndex].rows

        for (let i = 0; i < this.originList.length; i++) {
            if ((i === this.idsIndex && this.idsPoint.x > 500) || (this.idsPoint.x <= 500 && i !== this.idsIndex)) {
                cv.rotate(this.originList[i], this.originList[i], cv.ROTATE_90_COUNTERCLOCKWISE)
                cv.rotate(this.grayList[i], this.grayList[i], cv.ROTATE_90_COUNTERCLOCKWISE)
                cv.rotate(this.adaptiveList[i], this.adaptiveList[i], cv.ROTATE_90_COUNTERCLOCKWISE)
            } else {
                cv.rotate(this.originList[i], this.originList[i], cv.ROTATE_90_CLOCKWISE)
                cv.rotate(this.grayList[i], this.grayList[i], cv.ROTATE_90_CLOCKWISE)
                cv.rotate(this.adaptiveList[i], this.adaptiveList[i], cv.ROTATE_90_CLOCKWISE)
            }
        }

        if (this.idsPoint.x > 500) {
            this.idsPoint.x = this.idsMinY
            this.idsPoint.y = width - this.idsMaxX
        } else {
            this.idsPoint.x = height - this.idsMaxY
            this.idsPoint.y = this.idsMinX
        }

        console.log('ggggggggg')

        let rect = new cv.Rect(this.idsPoint.x + 386, this.idsPoint.y + 80, 580, 310)
        let sp1 = new cv.Point(rect.x, rect.y)
        let sp2 = new cv.Point(rect.x + rect.width, rect.y + rect.height)
        cv.rectangle(this.originList[this.idsIndex], sp1, sp2, [0, 0, 255, 255], 2)
        let phone = this.getPhone({
            rect,
            space: {x: 11, y: 11},
            size: {x: 30, y: 22},
        })
        let quesIndex = 1

        for (let i = 0; i < this.originList.length; i++) {
            let corners = new cv.MatVector();
            let nodes = new cv.Mat();
            let tmpInvMat = new cv.Mat()

            cv.morphologyEx(this.adaptiveList[i], this.adaptiveList[i], cv.MORPH_CLOSE, cv.Mat.ones(1, 30, cv.CV_8U));
            // cv.imshow('output', this.adaptiveList[i])

            cv.bitwise_not(this.adaptiveList[i], tmpInvMat);
            cv.findContours(tmpInvMat, corners, nodes, cv.RETR_LIST, cv.CHAIN_APPROX_SIMPLE)
            tmpInvMat.delete()


            let left_rectangles = []
            let right_rectangles = []

            for (let j = 0; j < corners.size(); j++) {
                let contour = corners.get(j)
                let rect = cv.boundingRect(contour)
                let minAreaRect = cv.minAreaRect(contour)
                contour.delete()

                let realSize = minAreaRect.size
                let realHeight = Math.min(realSize.width, realSize.height)
                let realWidth = Math.max(realSize.width, realSize.height)
                if (this.SL_MIN_WIDTH <= realWidth && realWidth <= this.SL_MAX_WIDTH && realHeight > this.SL_MIN_HEIGHT && realHeight < this.SL_MAX_HEIGHT) {
                    cv.drawContours(this.originList[i], corners, j, [255, 0, 0, 255], 2)
                    if (rect.x < width / 2) {
                        left_rectangles.push({
                            rect,
                            width: realWidth,
                            height: realHeight
                        })
                    } else {
                        right_rectangles.push({
                            rect,
                            width: realWidth,
                            height: realHeight
                        })
                    }
                }
            }
            corners.delete()
            nodes.delete()
            // cv.imshow('output', this.originList[i])

            for (let rectangles of [left_rectangles, right_rectangles]) {
                let [page_w, page_h] = [1450, 2290]
                let pageStart = rectangles.find(obj => obj.height <= 20 && obj.rect.y < 100)
                rectangles.sort((a, b) => a.rect.y - b.rect.y)
                if (pageStart == null) {
                    pageStart = rectangles[0]
                    pageStart.rect.y += 20
                }

                rectangles = rectangles.filter(obj => obj != pageStart && obj.height > 30 && obj.height < 40)
                for (let split of rectangles) {
                    let roiRect = new cv.Rect(pageStart.rect.x, pageStart.rect.y + 20, page_w, split.rect.y - 30 - pageStart.rect.y)
                    console.log(roiRect)
                    let image = this.toHsv(this.originList[i].roi(roiRect))
                    image_list.push({
                        id: quesIndex,
                        image
                    })
                    pageStart = split
                    pageStart.rect.y += 20
                    quesIndex += 1
                }

                let roiRect = new cv.Rect(pageStart.rect.x, pageStart.rect.y + 20, page_w, page_h - 30 - pageStart.rect.y)
                let image = this.toHsv(this.originList[i].roi(roiRect))
                image_list.push({
                    id: quesIndex,
                    image
                })
            }
        }

        return {
            status: "success",
            data: {
                phone,
                image_list
            }
        }
    }

    public static async initTemplate() {
        let tempGray = await Detect.readMatFromBase64('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABwAAAAcCAYAAAByDd+UAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAC9SURBVEhL7Y5JDsQgDAT5/6dJfCACU94yKKcpqQ/BvaT1j/kPLrTWNlUJEzTiKcJ1UGFGHnilkjcitlcKijzILyLCwQo6S/nlJTJn0B265/mKjBW8HnPwF7yu0qD2WD7B8uAgMd+1CMtzZFCkse5HBgnLg4OzYUZ7hgjLUxoU3vhmyoMZvK7lyzNW8HrcQW3OEOW3Fx2gEEE5kQbbKDikIc8QYf4+FVRkYV9uqChSROy4oWJShpzrIB8P9n4BbDAcY+LUK9IAAAAASUVORK5CYII=')
        cv.cvtColor(tempGray, tempGray, cv.COLOR_RGBA2GRAY, 0)
        return tempGray
    }

    public static detectFromCamara(mat, tempGray) {
        let resp = {}
        let points = []
        let grayImg = new cv.Mat()
        let cannyImg = new cv.Mat()
        resp.status = 'success'

        cv.cvtColor(mat, grayImg, cv.COLOR_RGBA2GRAY, 0)
        cv.GaussianBlur(grayImg, grayImg, new cv.Size(3, 3), 0, 0)
        cv.adaptiveThreshold(grayImg, grayImg, 255, cv.ADAPTIVE_THRESH_GAUSSIAN_C, cv.THRESH_BINARY, 15, 1)
        cv.Canny(grayImg, cannyImg, 10, 400)

        let contours = new cv.MatVector();
        let hierarchy = new cv.Mat();
        cv.findContours(grayImg, contours, hierarchy, cv.RETR_TREE, cv.CHAIN_APPROX_SIMPLE);

        try {
            for (let i = 0; i < contours.size(); i++) {
                let k = i
                let ic = 0
                while (hierarchy.intPtr(0, k)[2] !== -1) {
                    k = hierarchy.intPtr(0, k)[2]
                    ic++
                }
                if (ic == 4) {
                    cv.drawContours(mat, contours, i, [255, 0, 0, 255], 2)
                    let p = cv.moments(contours.get(i))
                    points.push({
                        x: p.m10 / p.m00,
                        y: p.m01 / p.m00
                    })
                }
            }

            // console.log(points)
            //

            if (points.length !== 4) {
                return resp
            }

            resp.strp = 1

            let four = Detect.fourPointsTransform(mat, points)

            let origin = Detect.resize(four, 248 * 4, 351 * 4)

            four.delete()

            four = new cv.Mat()
            cv.cvtColor(origin, four, cv.COLOR_RGBA2GRAY, 0)
            cv.adaptiveThreshold(four, four, 255, cv.ADAPTIVE_THRESH_MEAN_C, cv.THRESH_BINARY, 9, 1)

            resp.phone = Detect.getPhoneFromCamara(origin, four, {
                rect: {
                    x: 503,
                    y: 92,
                    width: 485,
                    height: 230
                }
            })

            let res = new cv.Mat()
            cv.matchTemplate(four, tempGray, res, cv.TM_CCOEFF_NORMED);
            // cv.imshow('output', four)
            resp.strp = 1.5

            const threshold = 0.7;
            const [w, h] = [tempGray.cols, tempGray.rows];

            const locationsData = res.data32F;
            console.log(locationsData)
            let locations = [];
            for (let i = 0; i < locationsData.length; i++) {
                if (locationsData[i] >= threshold) {
                    const x = i % res.cols;
                    const y = Math.floor(i / res.cols);
                    locations.push({
                        point: new cv.Point(x + w / 2, y + h / 2),
                        score: locationsData[i]
                    });
                }
            }


            resp.strp = 2
            locations.sort((a, b) => b.score - a.score);
            // 过滤距离小于 50 的重复点
            let filteredLocations = [];
            for (const loc of locations) {
                const isNearbyHigherScore = filteredLocations.some(filteredLoc => {
                    const dist = Detect.euclideanDistance(loc.point, filteredLoc.point);
                    return dist < 50 && filteredLoc.score >= loc.score;
                });
                if (!isNearbyHigherScore) {
                    filteredLocations.push(loc);
                }
            }

            resp.strp = 3
            filteredLocations = filteredLocations.map(obj => obj.point)
            filteredLocations = filteredLocations.filter(obj => obj .x > 100 && obj.x < 900 && obj.y < 1330 && obj.y > 100)
            filteredLocations = filteredLocations.sort((a, b) => b.x - a.x);

            for (let i = 0; i < filteredLocations.length; i++) {
                cv.circle(origin, filteredLocations[i], 5, [0, 255, 0, 255], 5)
            }

            console.log(filteredLocations)
            // cv.imshow('output', origin)

            resp.strp = 4

            if (filteredLocations.length === 21) {
                filteredLocations.sort((a, b) => b.y - a.y)
                filteredLocations.reverse()
                let rowList = []

                let index = 0
                rowList[0] = []
                rowList[0].push(filteredLocations[0])
                filteredLocations[0].index = 0
                for (let i = 1; i < filteredLocations.length; i++) {
                    filteredLocations[i].index = i;
                    if (filteredLocations[i].y - filteredLocations[i - 1].y > 50) {
                        index++
                        rowList[index] = []
                    }
                    rowList[index].push(filteredLocations[i])
                }

                for (let row of rowList) {
                    const avgY = row.reduce((sum, point) => sum + point.y, 0) / row.length;
                    // 添加第一个点,x坐标为5,y坐标为平均值
                    const point1 = new cv.Point(5, avgY);
                    row.unshift(point1);
                    // 添加第二个点,x坐标为four的宽度-5,y坐标和第一个点相同
                    const point2 = new cv.Point(four.cols - 5, avgY);
                    row.push(point2);
                    row.sort((a, b) => a.x - b.x);
                }

                // 添加最后一行点
                const lastRow = rowList[0].map(point => new cv.Point(point.x, four.rows - 5));
                rowList.push(lastRow);

                let cardList = []
                for (let i = 0; i < rowList.length - 1; i++) {
                    const currentRow = rowList[i];
                    const nextRow = rowList[i + 1];

                    for (let j = 0; j < 4; j++) {
                        const rect = [
                            currentRow[j],
                            currentRow[j + 1],
                            nextRow[j + 1],
                            nextRow[j]
                        ];

                        const rectImg = Detect.fourPointsTransform(origin, rect);
                        let canv = document.createElement('canvas');
                        let cardMat = new cv.Mat()
                        let cardGrayMat = new cv.Mat()
                        cv.cvtColor(rectImg, cardGrayMat, cv.COLOR_RGBA2GRAY, 0)
                        cv.adaptiveThreshold(cardGrayMat, cardMat, 255, cv.ADAPTIVE_THRESH_MEAN_C, cv.THRESH_BINARY, 33, 1)
                        let cardMatResized = Detect.resize(cardMat, 250, 150)
                        cardList.push(cardMatResized)
                        // cv.imshow(canv, cardMatResized)
                        document.body.appendChild(canv)
                        rectImg.delete()
                        cardGrayMat.delete()
                        cardMat.delete()
                    }
                }

                resp.answer = []
                let tmIndex = 0
                for (let card of cardList) {
                    for (let x = 0; x < 5; x++) {
                        tmIndex += 1
                        let ans = []
                        for (let y = 0; y < 4; y++) {
                            let startX = 16 + x * 45
                            let startY = 37 + y * 25
                            let w = 35
                            let h = 17
                            console.log(this.xingce_value, '0000')
                            if (Detect.calcRectWeight(card, new cv.Rect(startX, startY, w, h)) > this.xingce_value) {
                                ans.push('abcd'[y])
                            }
                        }

                        resp.answer.push({
                            "id": tmIndex,
                            "answer": ans
                        })
                    }
                }


                console.log(cardList)

            } else {
                resp.status = 'error'
                resp.message = '区域缺失，请确保没有处于强光/弱光状态，或者尽量保持答题卡纸面平整'
            }

            origin.delete()
            four.delete()
            res.delete()
        } catch (e) {
            console.error(e, e.stack)
            resp.status = 'error'
            resp.message = e.toString()
            return resp
        } finally {
            hierarchy.delete()
            contours.delete()
            grayImg.delete()
            cannyImg.delete()
        }

        return resp
    }

    private detectXcCard() {
        let width = this.originList[this.idsIndex].cols
        let height = this.originList[this.idsIndex].rows

        if (width > height) {
            if (this.idsPoint.x > 500) {
                this.idsPoint.x = this.idsMinY
                this.idsPoint.y = width - this.idsMaxX
                cv.rotate(this.originList[0], this.originList[0], cv.ROTATE_90_COUNTERCLOCKWISE)
                cv.rotate(this.grayList[0], this.grayList[0], cv.ROTATE_90_COUNTERCLOCKWISE)
                cv.rotate(this.adaptiveList[0], this.adaptiveList[0], cv.ROTATE_90_COUNTERCLOCKWISE)
            } else {
                this.idsPoint.x = height - this.idsMaxY
                this.idsPoint.y = this.idsMinX
                cv.rotate(this.originList[0], this.originList[0], cv.ROTATE_90_CLOCKWISE)
                cv.rotate(this.grayList[0], this.grayList[0], cv.ROTATE_90_CLOCKWISE)
                cv.rotate(this.adaptiveList[0], this.adaptiveList[0], cv.ROTATE_90_CLOCKWISE)
            }
        } else if (this.idsPoint.y > height / 2) {
            // 逆时针旋转180度
            this.idsPoint.x = width - this.idsMinX;
            this.idsPoint.y = height - this.idsMinY;
            cv.rotate(this.originList[0], this.originList[0], cv.ROTATE_180);
            cv.rotate(this.grayList[0], this.grayList[0], cv.ROTATE_180);
            cv.rotate(this.adaptiveList[0], this.adaptiveList[0], cv.ROTATE_180);
        }

        let tmpMat = Detect.resize(this.originList[0], 1080, 1920)
        this.originList[0].delete()
        this.originList[0] = tmpMat

        // console.log(tmpMat)
        // cv.imshow('output', this.originList[0])

        return Detect.detectFromCamara(this.originList[0], Detect.tempGray)
    }

    private detectInternal(imageList: any[]) {
        this.originList = imageList

        // 对传入的图片进行预处理
        this.preprocess(imageList)
        if (this.idsCount != 1) {
            return {
                status: 'error',
                message: this.idsCount === 0 ? '未找到标志' : '标志不唯一'
            }
        }

        if (this.idsIndex !== 0) {
            this.grayList.unshift(this.grayList.splice(this.idsIndex, 1)[0])
            this.adaptiveList.unshift(this.adaptiveList.splice(this.idsIndex, 1)[0])
            this.originList.unshift(this.originList.splice(this.idsIndex, 1)[0])
            this.idsIndex = 0
        }

        // 根据检测到的标志，处理类型不同的答题卡
        switch (this.ids) {
            case 0: return this.detectXcCard()
            case 1: return this.detectSlCard()
            default: return {
                status: "error",
                message: "未知标记:" + this.ids
            }
        }
    }

    // 唯一入口
    static detect(imageList: any[]): DetectResult {
        let obj = new Detect()

        try {
            return obj.detectInternal(imageList)
        } catch (e) {
            console.error(e.stack)
            return {
                status: "error",
                message: "处理异常：" + e.toString()
            }
        } finally {
            obj.clear()
        }
    }

    private clear() {
        this.dictionary.delete()
        this.grayList.forEach(obj => obj.delete())
        this.grayList = []

        this.adaptiveList.forEach(obj => obj.delete())
        this.adaptiveList = []

        this.originList.forEach(obj => obj.delete())
        this.originList = []
    }

    static async readMatFromBase64(imgStr: string) {
        let img = new Image()
        img.src = imgStr
        await new Promise(r => {
            img.onload = r
        })

        // @ts-ignore
        return cv.imread(img)
    }

    static async readMatFromFile(e: any, index?: number,type:any) {
        index = index || 0
        const fileInput = e.target || e;
        const file = fileInput.files ? fileInput.files[index] : e;
        if(type=='1'){
            this.xingce_value=0.5
        }
        console.log(file)
        if (file) {
            let fileReader = new FileReader()

            let imageData = await new Promise(r => {
                fileReader.onload = r
                fileReader.readAsDataURL(file)
            }) as any

            let img = new Image()
            img.src = imageData.target.result
            await new Promise(r => {
                img.onload = r
            })

            // @ts-ignore
            return cv.imread(img)
        }
    }
}



export default Detect
